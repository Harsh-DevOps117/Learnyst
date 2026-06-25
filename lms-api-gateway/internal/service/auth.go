package service

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"lms-api-gateway/internal/config"
	dtos "lms-api-gateway/internal/config/dto" // Adjust this path if your DTOs are elsewhere
	"lms-api-gateway/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db        *gorm.DB
	redisRepo *config.RedisRepo
}

func NewAuthService(db *gorm.DB, redisRepo *config.RedisRepo) *AuthService {
	return &AuthService{
		db:        db,
		redisRepo: redisRepo,
	}
}

// ==========================================
// PHASE 1: INITIATE SIGNUP (Form -> Redis -> SMS)
// ==========================================
func (a *AuthService) InitiateSignup(ctx context.Context, user *dtos.SignupRequest) error {
	// 1. Check if user already exists
	var count int64
	err := a.db.Model(&models.User{}).Where("email = ? OR phone = ?", user.Email, user.Phone).Count(&count).Error
	if err != nil {
		config.Log.Error("Database error while checking user existence: " + err.Error())
		return errors.New("internal server error")
	}
	if count > 0 {
		return errors.New("user with same email or phone number already exists")
	}

	// 2. Marshal user data to JSON
	userDataJSON, err := json.Marshal(user)
	if err != nil {
		config.Log.Error("Failed to marshal user data: " + err.Error())
		return errors.New("failed to process registration data")
	}

	// 3. Store in Redis Limbo (10 mins)
	if err := a.redisRepo.StorePendingUser(ctx, user.Phone, userDataJSON); err != nil {
		config.Log.Error("Redis Limbo storage failed: " + err.Error())
		return errors.New("failed to cache pending registration")
	}

	// 4. Generate & Store 6-digit OTP
	otp, err := config.GenerateOTP()
	if err != nil {
		config.Log.Error("OTP Generation failed: " + err.Error())
		return errors.New("failed to generate verification code")
	}

	if err := a.redisRepo.StoreOTP(ctx, user.Phone, otp); err != nil {
		config.Log.Error("Redis OTP storage failed: " + err.Error())
		return errors.New("failed to store verification code")
	}

	// 5. Send the SMS
	notifier := config.NewNotifier()
	config.Log.Info("DEVELOPMENT OTP for " + user.Phone + ": " + otp)
	if err := notifier.SendSMSOTP(user.Phone, otp); err != nil {
		config.Log.Error("SMS Dispatch failed: " + err.Error() + ". Bypassing for local development.")
		// We don't return an error here so local testing can proceed.
	}
	return nil
}

// ==========================================
// PHASE 1.5: RESEND OTP (With 60s Cooldown)
// ==========================================
func (a *AuthService) ResendSignupOTP(ctx context.Context, phone string) error {
	if a.redisRepo.IsOTPOnCooldown(ctx, phone) {
		return errors.New("please wait 60 seconds before requesting a new code")
	}

	_, err := a.redisRepo.GetPendingUser(ctx, phone)
	if err != nil {
		return errors.New("signup session expired. please fill out the registration form again")
	}

	newOTP, err := config.GenerateOTP()
	if err != nil {
		return errors.New("failed to generate new verification code")
	}

	if err := a.redisRepo.StoreOTP(ctx, phone, newOTP); err != nil {
		return errors.New("failed to store new verification code")
	}

	_ = a.redisRepo.SetOTPCooldown(ctx, phone)

	notifier := config.NewNotifier()
	config.Log.Info("DEVELOPMENT OTP for " + phone + ": " + newOTP)
	if err := notifier.SendSMSOTP(phone, newOTP); err != nil {
		config.Log.Error("SMS Dispatch failed: " + err.Error() + ". Bypassing for local development.")
		// We don't return an error here so local testing can proceed.
	}

	return nil
}

func (a *AuthService) CompleteSignup(ctx context.Context, phone string, submittedCode string) (*models.User, error) {

	savedOTP, err := a.redisRepo.GetOTP(ctx, phone)
	if err != nil || savedOTP != submittedCode {
		return nil, errors.New("invalid or expired verification code")
	}

	cachedData, err := a.redisRepo.GetPendingUser(ctx, phone)
	if err != nil {
		return nil, errors.New("registration session expired. please sign up again")
	}

	var req dtos.SignupRequest
	if err := json.Unmarshal(cachedData, &req); err != nil {
		return nil, errors.New("failed to parse cached registration data")
	}

	pepper := os.Getenv("APP_SECRET_PEPPER")
	pepperedPassword := req.Password + pepper
	hash, err := bcrypt.GenerateFromPassword([]byte(pepperedPassword), bcrypt.DefaultCost)
	if err != nil {
		config.Log.Error("Password hashing failed: " + err.Error())
		return nil, errors.New("failed to secure user password")
	}

	user := models.User{
		First_Name:     req.FirstName,
		Middle_Name:    req.MiddleName,
		Last_Name:      req.LastName,
		Email:          req.Email,
		Phone:          req.Phone,
		Location:       req.Location,
		RoleID:         req.RoleID,
		Password_Hash:  string(hash),
		Phone_Verified: true,
		Email_Verified: false,
		Is_Active:      true,
	}

	if err := a.db.Create(&user).Error; err != nil {
		config.Log.Error("Postgres insertion failed: " + err.Error())
		return nil, errors.New("failed to finalize user creation in database")
	}

	_ = a.redisRepo.DeletePendingUser(ctx, phone)
	_ = a.redisRepo.DeleteOTP(ctx, phone)
	go func() {
		_ = a.DispatchEmailVerification(context.Background(), user.Email)
	}()

	return &user, nil
}

func (a *AuthService) DispatchEmailVerification(ctx context.Context, email string) error {
	token := uuid.New().String()

	if err := a.redisRepo.StoreVerificationToken(ctx, token, email); err != nil {
		return err
	}

	notifier := config.NewNotifier()
	return notifier.SendEmailVerificationLink(email, token)
}

func (a *AuthService) VerifyEmailLinkClick(ctx context.Context, token string) error {
	email, err := a.redisRepo.GetEmailFromToken(ctx, token)
	if err != nil {
		return errors.New("verification link is invalid or has expired")
	}

	result := a.db.Model(&models.User{}).Where("email = ?", email).Update("email_verified", true)
	if result.Error != nil {
		return errors.New("database error while verifying email")
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	_ = a.redisRepo.DeleteVerificationToken(ctx, token)

	return nil
}

// ==========================================
// LOGIN
// ==========================================
func (a *AuthService) Login(ctx context.Context, req *dtos.LoginRequest) (*dtos.AuthResponse, error) {
	var user models.User
	if err := a.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, errors.New("database error")
	}

	pepper := os.Getenv("APP_SECRET_PEPPER")
	pepperedPassword := req.Password + pepper
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password_Hash), []byte(pepperedPassword)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !user.Is_Active {
		return nil, errors.New("user account is inactive")
	}

	accessToken, refreshToken, err := config.GenerateTokens(user.ID, user.RoleID)
	if err != nil {
		config.Log.Error("Token generation failed: " + err.Error())
		return nil, errors.New("failed to generate authentication tokens")
	}

	// Store refresh token in db
	rt := models.Refresh_Token{
		Token:       refreshToken,
		UserID:      user.ID,
		Device_Name: req.DeviceName,
		IP_Address:  "", // IP will be captured later if needed
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	}
	a.db.Create(&rt)

	return &dtos.AuthResponse{
		Message:      "Login successful",
		UserID:       user.ID,
		RoleID:       uint(user.RoleID),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(15 * time.Minute).Unix(),
	}, nil
}

// ==========================================
// PASSWORD RESET
// ==========================================
func (a *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	var user models.User
	if err := a.db.Where("email = ?", email).First(&user).Error; err != nil {
		// Do not leak if user exists or not, but for testing we could return error
		return nil
	}

	token := uuid.New().String()
	if err := a.redisRepo.StoreVerificationToken(ctx, token, email); err != nil {
		return err
	}

	notifier := config.NewNotifier()
	return notifier.SendPasswordResetLink(email, token)
}

func (a *AuthService) CompletePasswordReset(ctx context.Context, token string, newPassword string) error {
	email, err := a.redisRepo.GetEmailFromToken(ctx, token)
	if err != nil {
		return errors.New("reset link is invalid or has expired")
	}

	pepper := os.Getenv("APP_SECRET_PEPPER")
	pepperedPassword := newPassword + pepper
	hash, err := bcrypt.GenerateFromPassword([]byte(pepperedPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to secure user password")
	}

	result := a.db.Model(&models.User{}).Where("email = ?", email).Update("password_hash", string(hash))
	if result.Error != nil {
		return errors.New("database error while resetting password")
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	_ = a.redisRepo.DeleteVerificationToken(ctx, token)
	return nil
}

// ==========================================
// CHANGE PASSWORD
// ==========================================
func (a *AuthService) ChangePassword(ctx context.Context, userID string, oldPassword string, newPassword string) error {
	var user models.User
	if err := a.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	pepper := os.Getenv("APP_SECRET_PEPPER")
	pepperedOldPassword := oldPassword + pepper
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password_Hash), []byte(pepperedOldPassword)); err != nil {
		return errors.New("incorrect old password")
	}

	pepperedNewPassword := newPassword + pepper
	hash, err := bcrypt.GenerateFromPassword([]byte(pepperedNewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to secure new password")
	}

	if err := a.db.Model(&user).Update("password_hash", string(hash)).Error; err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

// ==========================================
// CHANGE EMAIL
// ==========================================
func (a *AuthService) InitiateChangeEmail(ctx context.Context, userID string, newEmail string) error {
	var user models.User
	if err := a.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return errors.New("user not found")
	}

	// Check if email is already in use
	var count int64
	a.db.Model(&models.User{}).Where("email = ?", newEmail).Count(&count)
	if count > 0 {
		return errors.New("email is already in use by another account")
	}

	// For email change, we'll store JSON like {"user_id": userID, "new_email": newEmail} against the token
	token := uuid.New().String()
	payload := map[string]string{"user_id": userID, "new_email": newEmail}
	payloadBytes, _ := json.Marshal(payload)
	
	// Re-using the same store verification token method for payload storage since it's just a string setter
	if err := a.redisRepo.StoreVerificationToken(ctx, token, string(payloadBytes)); err != nil {
		return errors.New("failed to cache verification token")
	}

	notifier := config.NewNotifier()
	return notifier.SendEmailChangeVerificationLink(newEmail, token)
}

func (a *AuthService) VerifyChangeEmail(ctx context.Context, token string) error {
	payloadStr, err := a.redisRepo.GetEmailFromToken(ctx, token)
	if err != nil {
		return errors.New("verification link is invalid or has expired")
	}

	var payload map[string]string
	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
		return errors.New("invalid payload data")
	}

	userID, ok1 := payload["user_id"]
	newEmail, ok2 := payload["new_email"]
	if !ok1 || !ok2 {
		return errors.New("invalid payload data format")
	}

	result := a.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"email":          newEmail,
		"email_verified": true, // assume it's verified now
	})

	if result.Error != nil {
		return errors.New("database error while changing email")
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}

	_ = a.redisRepo.DeleteVerificationToken(ctx, token)
	return nil
}
