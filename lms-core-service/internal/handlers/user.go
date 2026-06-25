package handlers

import (
	"net/http"
	"lms-core-service/internal/database"
	"lms-core-service/internal/models"

	"github.com/gin-gonic/gin"
)

func GetUserHandler(c *gin.Context) {
	userId := c.Param("id")

	db := database.DB
	var user models.User

	if err := db.Select("id, first_name, last_name, email, phone, location, role_id, profile_image_url").Where("id = ?", userId).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	type UserResponse struct {
		models.User
		AcademySlug string `json:"academy_slug,omitempty"`
		TenantID    uint   `json:"tenant_id,omitempty"`
	}

	resp := UserResponse{User: user}

	// Check if user owns a tenant, regardless of their current role_id
	var tenant models.Tenant
	if err := db.Select("id, academy_slug").Where("owner_user_id = ?", user.ID).First(&tenant).Error; err == nil {
		resp.AcademySlug = tenant.AcademySlug
		resp.TenantID = tenant.ID
	}

	c.JSON(http.StatusOK, resp)
}

type UpdateUserProfileRequest struct {
	ProfileImageUrl string `json:"profile_image_url"`
}

func UpdateUserProfileHandler(c *gin.Context) {
	userId := c.Param("id")
	var req UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Model(&models.User{}).Where("id = ?", userId).
		Updates(map[string]interface{}{
			"profile_image_url": req.ProfileImageUrl,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User profile updated successfully"})
}
