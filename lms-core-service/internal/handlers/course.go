package handlers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"lms-core-service/internal/database"
	"lms-core-service/internal/kafka"
	lmsmodels "lms-core-service/internal/models"
	"lms-core-service/internal/storage"
	"net/http"
)

type CreateCourseRequest struct {
	TenantID    uint    `json:"tenant_id" binding:"required"`
	Title       string  `json:"title" binding:"required"`
	Slug        string  `json:"slug" binding:"required"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Level       string  `json:"level"`
	Price       float64 `json:"price"`
	CreatedBy   uint    `json:"created_by" binding:"required"`
}

func CreateCourseHandler(c *gin.Context) {
	var req CreateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var staff lmsmodels.TenantStaff
	if err := database.DB.Where("tenant_id = ? AND user_id = ?", req.TenantID, req.CreatedBy).First(&staff).Error; err != nil {
		var tenant lmsmodels.Tenant
		if err := database.DB.Where("id = ? AND owner_user_id = ?", req.TenantID, req.CreatedBy).First(&tenant).Error; err == nil {
			staff = lmsmodels.TenantStaff{
				TenantID: req.TenantID,
				UserID:   req.CreatedBy,
				RoleID:   1,
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
		CreatedBy:   staff.ID,
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
	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.ThumbnailUrl != "" {
		updates["thumbnail_url"] = req.ThumbnailUrl
	}
	if req.Price != 0 {
		updates["price"] = req.Price
	}
	if req.IsPublished != nil {
		updates["is_published"] = *req.IsPublished
	}
	if len(updates) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No updates provided"})
		return
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
	NoteUrl         string `json:"note_url"`
	DurationSeconds int    `json:"duration_seconds"`
}

func CreateCourseVideoHandler(c *gin.Context) {
	courseId := c.Param("id")
	var req CreateVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var cID uint
	fmt.Sscan(courseId, &cID)
	var staff lmsmodels.TenantStaff
	if err := database.DB.Where("user_id = ? AND tenant_id = ?", req.TeacherID, req.TenantID).First(&staff).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is not recognized as staff for this academy"})
		return
	}
	var tenant lmsmodels.Tenant
	if err := database.DB.Preload("Owner").First(&tenant, req.TenantID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tenant details"})
		return
	}
	video := lmsmodels.CourseVideo{
		TenantID:        req.TenantID,
		CourseID:        cID,
		TeacherID:       staff.ID,
		Title:           req.Title,
		Description:     req.Description,
		VideoUrl:        req.VideoUrl,
		NoteUrl:         req.NoteUrl,
		DurationSeconds: req.DurationSeconds,
		IsPublished:     true,
	}
	if err := database.DB.Create(&video).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create video: " + err.Error()})
		return
	}
	go func() {
		_ = kafka.PublishVideoUploaded(kafka.VideoUploadedEvent{
			TenantID: video.TenantID,
			CourseID: video.CourseID,
			VideoID:  video.ID,
			RawUrl:   video.VideoUrl,
			Slug:     tenant.AcademySlug,
			Phone:    tenant.Owner.Phone,
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
	var staff lmsmodels.TenantStaff
	if err := database.DB.Where("user_id = ? AND tenant_id = ?", req.TeacherID, req.TenantID).First(&staff).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is not recognized as staff for this academy"})
		return
	}
	assignment := lmsmodels.Assignment{
		TenantID:      req.TenantID,
		CourseID:      cID,
		TeacherID:     staff.ID,
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
func DeleteCourseHandler(c *gin.Context) {
	id := c.Param("id")
	var course lmsmodels.Course
	if err := database.DB.First(&course, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}
	if course.ThumbnailUrl != "" {
		_ = storage.DeleteBlob(context.Background(), course.ThumbnailUrl)
	}
	// 2. Fetch and Delete Videos (blob + db record)
	var videos []lmsmodels.CourseVideo
	database.DB.Where("course_id = ?", course.ID).Find(&videos)
	for _, video := range videos {
		if video.VideoUrl != "" {
			_ = storage.DeleteBlob(context.Background(), video.VideoUrl)
		}
		database.DB.Delete(&video)
	}
	// 3. Fetch and Delete Assignments (blob + db record)
	var assignments []lmsmodels.Assignment
	database.DB.Where("course_id = ?", course.ID).Find(&assignments)
	for _, assignment := range assignments {
		if assignment.AttachmentUrl != "" {
			_ = storage.DeleteBlob(context.Background(), assignment.AttachmentUrl)
		}
		// Also delete submissions before deleting the assignment
		database.DB.Where("assignment_id = ?", assignment.ID).Delete(&lmsmodels.AssignmentSubmission{})
		database.DB.Delete(&assignment)
	}
	// 4. Delete other dependent relations to satisfy foreign key constraints
	database.DB.Where("course_id = ?", course.ID).Delete(&lmsmodels.Invitation{})
	database.DB.Where("course_id = ?", course.ID).Delete(&lmsmodels.CourseTeacher{})
	database.DB.Where("course_id = ?", course.ID).Delete(&lmsmodels.CourseModule{})
	database.DB.Where("course_id = ?", course.ID).Delete(&lmsmodels.Note{})
	database.DB.Where("course_id = ?", course.ID).Delete(&lmsmodels.Coupon{})
	// 5. Handle Batches (which have enrollments and classes)
	var batches []lmsmodels.Batch
	database.DB.Where("course_id = ?", course.ID).Find(&batches)
	for _, batch := range batches {
		database.DB.Where("batch_id = ?", batch.ID).Delete(&lmsmodels.Enrollment{})
		var classes []lmsmodels.Class
		database.DB.Where("batch_id = ?", batch.ID).Find(&classes)
		for _, class := range classes {
			database.DB.Where("class_id = ?", class.ID).Delete(&lmsmodels.ClassAttendance{})
			database.DB.Delete(&class)
		}
		database.DB.Delete(&batch)
	}
	// 6. Delete the course itself
	if err := database.DB.Delete(&course).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete course from database: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Course and associated assets deleted successfully"})
}
