package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"lms-core-service/internal/database"
	"lms-core-service/internal/kafka"
	lmsmodels "lms-core-service/internal/models"
)

type CreateCourseRequest struct {
	TenantID     uint    `json:"tenant_id" binding:"required"`
	Title        string  `json:"title" binding:"required"`
	Slug         string  `json:"slug" binding:"required"`
	Description  string  `json:"description"`
	Category     string  `json:"category"`
	Level        string  `json:"level"`
	Price        float64 `json:"price"`
	CreatedBy    uint    `json:"created_by" binding:"required"`
}

func CreateCourseHandler(c *gin.Context) {
	var req CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure the user is a staff member of the tenant
	var staff lmsmodels.TenantStaff
	if err := database.DB.Where("tenant_id = ? AND user_id = ?", req.TenantID, req.CreatedBy).First(&staff).Error; err != nil {
		// Staff record not found. Check if the user is the tenant owner.
		var tenant lmsmodels.Tenant
		if err := database.DB.Where("id = ? AND owner_user_id = ?", req.TenantID, req.CreatedBy).First(&tenant).Error; err == nil {
			// User is the owner. Auto-create a TenantStaff record for them so they can act as a creator.
			staff = lmsmodels.TenantStaff{
				TenantID: req.TenantID,
				UserID:   req.CreatedBy,
				RoleID:   1, // Assuming 1 is Admin/Owner role
				IsActive: true,
			}
			if createErr := database.DB.Create(&staff).Error; createErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create staff record for owner"})
				return
			}
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "User is not authorized to create course for this tenant"})
			return
		}
	}

	course := lmsmodels.Course{
		TenantID:    req.TenantID,
		Title:       req.Title,
		Slug:        req.Slug,
		Description: req.Description,
		Category:    req.Category,
		Level:       req.Level,
		Price:       req.Price,
		CreatedBy:   staff.ID, // Use the resolved TenantStaff ID
		IsPublished: false,
	}

	if err := database.DB.Create(&course).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create course"})
		return
	}

	c.JSON(http.StatusOK, course)
}

type UpdateCourseRequest struct {
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	ThumbnailUrl string  `json:"thumbnail_url"`
	Price        float64 `json:"price"`
	IsPublished  *bool   `json:"is_published"`
}

func UpdateCourseHandler(c *gin.Context) {
	id := c.Param("id")
	var req UpdateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"title":         req.Title,
		"description":   req.Description,
		"thumbnail_url": req.ThumbnailUrl,
		"price":         req.Price,
	}

	if req.IsPublished != nil {
		updates["is_published"] = *req.IsPublished
	}

	if err := database.DB.Model(&lmsmodels.Course{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update course"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Course updated successfully"})
}

type CreateVideoRequest struct {
	TenantID        uint   `json:"tenant_id" binding:"required"`
	TeacherID       uint   `json:"teacher_id" binding:"required"`
	Title           string `json:"title" binding:"required"`
	Description     string `json:"description"`
	VideoUrl        string `json:"video_url" binding:"required"`
	DurationSeconds int    `json:"duration_seconds"`
}

func CreateCourseVideoHandler(c *gin.Context) {
	courseId := c.Param("id")
	var req CreateVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parsing courseId string to uint
	var cID uint
	fmt.Sscan(courseId, &cID)

	video := lmsmodels.CourseVideo{
		TenantID:        req.TenantID,
		CourseID:        cID,
		TeacherID:       req.TeacherID,
		Title:           req.Title,
		Description:     req.Description,
		VideoUrl:        req.VideoUrl,
		DurationSeconds: req.DurationSeconds,
		IsPublished:     true,
	}

	if err := database.DB.Create(&video).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create video: " + err.Error()})
		return
	}

	// Publish Video Uploaded Event for Adaptive Streaming Processing
	go func() {
		_ = kafka.PublishVideoUploaded(kafka.VideoUploadedEvent{
			TenantID: video.TenantID,
			CourseID: video.CourseID,
			VideoID:  video.ID,
			RawUrl:   video.VideoUrl,
		})
	}()

	c.JSON(http.StatusOK, video)
}

type CreateAssignmentRequest struct {
	TenantID      uint   `json:"tenant_id" binding:"required"`
	TeacherID     uint   `json:"teacher_id" binding:"required"`
	Title         string `json:"title" binding:"required"`
	Description   string `json:"description"`
	AttachmentUrl string `json:"attachment_url"`
	MaxMarks      int    `json:"max_marks"`
}

func CreateAssignmentHandler(c *gin.Context) {
	courseId := c.Param("id")
	var req CreateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var cID uint
	fmt.Sscan(courseId, &cID)

	assignment := lmsmodels.Assignment{
		TenantID:      req.TenantID,
		CourseID:      cID,
		TeacherID:     req.TeacherID,
		Title:         req.Title,
		Description:   req.Description,
		AttachmentUrl: req.AttachmentUrl,
		MaxMarks:      req.MaxMarks,
	}

	if err := database.DB.Create(&assignment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create assignment"})
		return
	}

	c.JSON(http.StatusOK, assignment)
}
