package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	dtos "lms-api-gateway/internal/config/dto"
	"lms-api-gateway/internal/models"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

var ErrCredentials = errors.New("Email or password is incorrect.")
var ErrSession = errors.New("Your session has expired. Please sign in again.")

type AuthService struct{ DB *gorm.DB }

func NewAuthService(db *gorm.DB) *AuthService { return &AuthService{DB: db} }
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
func randomToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	return hex.EncodeToString(b), err
}
func (a *AuthService) Register(req dtos.SignupRequest) (*models.User, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	if req.FirstName == "" || req.LastName == "" || len([]byte(req.Password)) > 72 {
		return nil, errors.New("Enter your name and a password of 8–72 bytes.")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := models.User{FirstName: req.FirstName, LastName: req.LastName, Email: req.Email, PasswordHash: string(hash), RoleID: req.RoleID, IsActive: true}
	if err := a.DB.Create(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.New("This email already has an account. Please sign in.")
		}
		return nil, errors.New("Could not create your account. Please try again.")
	}
	return &user, nil
}
func (a *AuthService) Login(req dtos.LoginRequest) (*models.User, error) {
	var user models.User
	err := a.DB.Where("email = ?", strings.ToLower(strings.TrimSpace(req.Email))).First(&user).Error
	if err != nil {
		// Keep the expensive password comparison for unknown accounts too.
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"), []byte(req.Password))
		return nil, ErrCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil || !user.IsActive {
		return nil, ErrCredentials
	}
	return &user, nil
}
func (a *AuthService) NewSession(userID uint) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}
	session := models.AuthSession{ID: HashToken(token), UserID: userID, ExpiresAt: time.Now().Add(7 * 24 * time.Hour)}
	// Opportunistic cleanup keeps expired credentials out of the database.
	a.DB.Where("expires_at <= ?", time.Now()).Delete(&models.AuthSession{})
	return token, a.DB.Create(&session).Error
}
func (a *AuthService) Session(token string) (*models.User, error) {
	if len(token) != 64 {
		return nil, ErrSession
	}
	var session models.AuthSession
	if a.DB.Where("id = ? AND expires_at > ?", HashToken(token), time.Now()).First(&session).Error != nil {
		return nil, ErrSession
	}
	var user models.User
	if a.DB.Where("id = ? AND is_active = ?", session.UserID, true).First(&user).Error != nil {
		return nil, ErrSession
	}
	return &user, nil
}
func (a *AuthService) Logout(token string) error {
	return a.DB.Where("id = ?", HashToken(token)).Delete(&models.AuthSession{}).Error
}
func (a *AuthService) RequestReset(email string) error {
	var user models.User
	if a.DB.Where("email = ? AND is_active = ?", strings.ToLower(strings.TrimSpace(email)), true).First(&user).Error != nil {
		return nil
	}
	token, err := randomToken()
	if err != nil {
		return err
	}
	reset := models.PasswordReset{ID: HashToken(token), UserID: user.ID, ExpiresAt: time.Now().Add(30 * time.Minute)}
	if err := a.DB.Create(&reset).Error; err != nil {
		return err
	}
	base := strings.TrimRight(os.Getenv("FRONTEND_URL"), "/")
	if base == "" {
		base = "http://localhost:3005"
	}
	body := "To: " + user.Email + "\r\nFrom: Learnyst <hello@learnyst.local>\r\nSubject: Reset your Learnyst password\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nReset your password: " + base + "/reset-password?token=" + token + "\r\n\r\nThis link expires in 30 minutes and can only be used once. If you did not request it, ignore this message."
	addr := os.Getenv("SMTP_ADDR")
	if addr == "" {
		addr = "127.0.0.1:1025"
	}
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		a.DB.Delete(&reset)
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	if err = client.Mail("hello@learnyst.local"); err != nil {
		return err
	}
	if err = client.Rcpt(user.Email); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write([]byte(body)); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
func (a *AuthService) ResetPassword(token, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("Password must be 8–72 bytes.")
	}
	return a.DB.Transaction(func(tx *gorm.DB) error {
		var reset models.PasswordReset
		if tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND expires_at > ?", HashToken(token), time.Now()).First(&reset).Error != nil {
			return errors.New("This reset link has expired or already been used.")
		}
		if err := tx.Model(&models.User{}).Where("id = ?", reset.UserID).Update("password_hash", string(hash)).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", reset.UserID).Delete(&models.AuthSession{}).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", reset.UserID).Delete(&models.PasswordReset{}).Error
	})
}
func (a *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	if a.DB.First(&user, userID).Error != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)) != nil {
		return errors.New("Current password is incorrect.")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("Password must be 8–72 bytes.")
	}
	return a.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Update("password_hash", string(hash)).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", userID).Delete(&models.AuthSession{}).Error
	})
}
