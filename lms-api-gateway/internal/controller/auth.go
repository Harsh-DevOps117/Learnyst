package controller

import (
	"github.com/gin-gonic/gin"
	dtos "lms-api-gateway/internal/config/dto"
	"lms-api-gateway/internal/models"
	"lms-api-gateway/internal/service"
	"log"
	"net/http"
	"os"
	"strings"
)

const SessionCookie = "learnyst_session"

type AuthController struct{ authService *service.AuthService }

func NewAuthController(s *service.AuthService) *AuthController {
	return &AuthController{authService: s}
}
func setCookie(c *gin.Context, token string, age int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(SessionCookie, token, age, "/", "", os.Getenv("COOKIE_SECURE") == "true", true)
	c.Header("Cache-Control", "no-store")
}
func respondUser(c *gin.Context, u *models.User, code int) {
	c.Header("Cache-Control", "no-store")
	c.JSON(code, gin.H{"user": u, "user_id": u.ID, "role_id": u.RoleID})
}
func (ac *AuthController) signIn(c *gin.Context, u *models.User, code int) {
	token, err := ac.authService.NewSession(u.ID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Could not start your session."})
		return
	}
	old, _ := c.Cookie(SessionCookie)
	if old != "" {
		_ = ac.authService.Logout(old)
	}
	setCookie(c, token, 7*24*3600)
	respondUser(c, u, code)
}
func (ac *AuthController) Register(c *gin.Context) {
	var req dtos.SignupRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"error": "Enter your first and last name, a valid email, a password of 8–72 characters, and a valid account type."})
		return
	}
	user, err := ac.authService.Register(req)
	if err != nil {
		c.JSON(409, gin.H{"error": err.Error()})
		return
	}
	ac.signIn(c, user, 201)
}
func (ac *AuthController) Login(c *gin.Context) {
	var req dtos.LoginRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"error": "Enter a valid email and password."})
		return
	}
	user, err := ac.authService.Login(req)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}
	ac.signIn(c, user, 200)
}
func (ac *AuthController) Session(c *gin.Context) {
	token, _ := c.Cookie(SessionCookie)
	user, err := ac.authService.Session(token)
	if err != nil {
		setCookie(c, "", -1)
		c.JSON(401, gin.H{"error": service.ErrSession.Error()})
		return
	}
	respondUser(c, user, 200)
}
func (ac *AuthController) Logout(c *gin.Context) {
	token, _ := c.Cookie(SessionCookie)
	if err := ac.authService.Logout(token); err != nil {
		c.JSON(503, gin.H{"error": "Could not sign out. Please try again."})
		return
	}
	setCookie(c, "", -1)
	c.JSON(200, gin.H{"message": "Signed out."})
}
func (ac *AuthController) RequestPasswordReset(c *gin.Context) {
	var req dtos.ResetPasswordRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"error": "Enter a valid email address."})
		return
	}
	// The same response for known and unknown emails prevents account enumeration.
	if err := ac.authService.RequestReset(req.Email); err != nil {
		log.Print("Password reset email could not be delivered")
	}
	c.JSON(200, gin.H{"message": "If this email has an account, a reset link has been sent. Check your local inbox."})
}
func (ac *AuthController) CompletePasswordReset(c *gin.Context) {
	var req dtos.CompleteResetPasswordRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"error": "Enter a valid reset token and a password of 8–72 characters."})
		return
	}
	if err := ac.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	setCookie(c, "", -1)
	c.JSON(200, gin.H{"message": "Password updated. Please sign in again."})
}
func (ac *AuthController) ChangePassword(c *gin.Context) {
	token, _ := c.Cookie(SessionCookie)
	user, err := ac.authService.Session(token)
	if err != nil {
		c.JSON(401, gin.H{"error": service.ErrSession.Error()})
		return
	}
	var req dtos.ChangePasswordRequest
	if c.ShouldBindJSON(&req) != nil || len(strings.TrimSpace(req.NewPassword)) < 8 {
		c.JSON(400, gin.H{"error": "Enter your current password and a new password of 8–72 characters."})
		return
	}
	if err := ac.authService.ChangePassword(user.ID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	setCookie(c, "", -1)
	c.JSON(200, gin.H{"message": "Password changed. Please sign in again."})
}
