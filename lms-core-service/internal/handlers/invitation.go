package handlers
import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"lms-core-service/internal/database"
	"lms-core-service/internal/mail"
	"lms-core-service/internal/models"
)
type SendInvitationRequest struct {
	Email    string `json:"email" binding:"required,email"`
	RoleID   uint   `json:"role_id" binding:"required"`
	CourseID *uint  `json:"course_id"`
}
type AcceptInvitationRequest struct {
	Token  string `json:"token" binding:"required"`
	UserID uint   `json:"user_id" binding:"required"` 
}
func generateSecureToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "fallback-token-" + time.Now().String()
	}
	return hex.EncodeToString(bytes)
}
func SendInvitationHandler(c *gin.Context) {
	tenantID := c.Param("tenant_id")
	var req SendInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var tenant models.Tenant
	if err := database.DB.First(&tenant, tenantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}
	domain := tenant.CustomDomain
	if domain == "" {
		domain = tenant.AcademySlug + ".learnyst.com" 
	}
	if req.CourseID != nil {
		var course models.Course
		if err := database.DB.Where("id = ? AND tenant_id = ?", *req.CourseID, tenantID).First(&course).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID for this tenant"})
			return
		}
	}
	token := generateSecureToken()
	expiry := time.Now().Add(7 * 24 * time.Hour)
	invitation := models.Invitation{
		TenantID:    tenant.ID,
		Email:       req.Email,
		RoleID:      req.RoleID,
		CourseID:    req.CourseID,
		InviteToken: token,
		ExpiresAt:   &expiry,
	}
	if err := database.DB.Create(&invitation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create invitation"})
		return
	}
	go func() {
		err := mail.SendInvitationEmail(req.Email, tenant.AcademyName, domain, token)
		if err != nil {
			println("Failed to send email:", err.Error())
		}
	}()
	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation sent successfully",
		"invitation_id": invitation.ID,
		"invite_token": token,
	})
}
func AcceptInvitationHandler(c *gin.Context) {
	var req AcceptInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var invitation models.Invitation
	if err := database.DB.Where("invite_token = ? AND accepted = ?", req.Token, false).First(&invitation).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or already accepted invitation token"})
		return
	}
	if invitation.ExpiresAt != nil && invitation.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invitation token has expired"})
		return
	}
	var user models.User
	if err := database.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}
	tx := database.DB.Begin()
	invitation.Accepted = true
	if err := tx.Save(&invitation).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update invitation"})
		return
	}
	staff := models.TenantStaff{
		TenantID: invitation.TenantID,
		UserID:   user.ID,
		RoleID:   invitation.RoleID,
		IsActive: true,
	}
	if err := tx.Create(&staff).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign staff role. They might already be staff."})
		return
	}
	if invitation.CourseID != nil {
		courseTeacher := models.CourseTeacher{
			TenantID:  invitation.TenantID,
			CourseID:  *invitation.CourseID,
			TeacherID: staff.ID,
		}
		if err := tx.Create(&courseTeacher).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to auto-assign course."})
			return
		}
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation accepted successfully",
		"tenant_id": invitation.TenantID,
		"course_assigned": invitation.CourseID != nil,
	})
}
