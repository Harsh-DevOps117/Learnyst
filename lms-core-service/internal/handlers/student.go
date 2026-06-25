package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"lms-core-service/internal/database"
	lmsmodels "lms-core-service/internal/models"
)

type TrackVideoProgressRequest struct {
	StudentID         uint `json:"student_id" binding:"required"`
	VideoID           uint `json:"video_id" binding:"required"`
	WatchTimeSeconds  int  `json:"watch_time_seconds"`
	CompletionPercent int  `json:"completion_percent"`
}

func TrackVideoProgressHandler(c *gin.Context) {
	var req TrackVideoProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var progress lmsmodels.VideoProgress
	err := database.DB.Where("student_id = ? AND video_id = ?", req.StudentID, req.VideoID).First(&progress).Error
	
	now := time.Now()
	if err != nil {
		// Create new
		progress = lmsmodels.VideoProgress{
			StudentID:         req.StudentID,
			VideoID:           req.VideoID,
			WatchTimeSeconds:  req.WatchTimeSeconds,
			CompletionPercent: req.CompletionPercent,
			LastWatchedAt:     &now,
		}
		if err := database.DB.Create(&progress).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to track progress"})
			return
		}
	} else {
		// Update existing
		updates := map[string]interface{}{
			"last_watched_at": now,
		}
		if req.WatchTimeSeconds > progress.WatchTimeSeconds {
			updates["watch_time_seconds"] = req.WatchTimeSeconds
		}
		if req.CompletionPercent > progress.CompletionPercent {
			updates["completion_percent"] = req.CompletionPercent
		}
		database.DB.Model(&progress).Updates(updates)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Progress tracked successfully", "progress": progress})
}

func ConsumeCourseHandler(c *gin.Context) {
	courseId := c.Param("id")
	studentId := c.Query("student_id")

	if studentId == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "student_id is required to consume course"})
		return
	}

	// 1. Security Check: Verify Student is enrolled in a batch for this course
	var count int64
	err := database.DB.Table("enrollments").
		Joins("JOIN batches ON batches.id = enrollments.batch_id").
		Where("batches.course_id = ? AND enrollments.student_id = ?", courseId, studentId).
		Count(&count).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify enrollment"})
		return
	}

	if count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. You are not enrolled in this course."})
		return
	}

	// 2. Fetch all videos for this course
	var videos []lmsmodels.CourseVideo
	if err := database.DB.Where("course_id = ?", courseId).Order("position ASC").Find(&videos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch course videos"})
		return
	}

	// 3. Fetch all assignments for this course
	var assignments []lmsmodels.Assignment
	if err := database.DB.Where("course_id = ?", courseId).Find(&assignments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch course assignments"})
		return
	}

	// 4. Fetch all notes for this course
	var notes []lmsmodels.Note
	if err := database.DB.Where("course_id = ?", courseId).Find(&notes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch course notes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"videos":      videos,
		"assignments": assignments,
		"notes":       notes,
	})
}

type CreateDoubtRequest struct {
	TenantID      uint   `json:"tenant_id" binding:"required"`
	StudentID     uint   `json:"student_id" binding:"required"`
	CourseID      *uint  `json:"course_id"`
	ReferenceType string `json:"reference_type"`
	ReferenceID   *uint  `json:"reference_id"`
	QuestionText  string `json:"question_text" binding:"required"`
}

func CreateDoubtQueryHandler(c *gin.Context) {
	var req CreateDoubtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// AI summary logic
	aiResponse := fmt.Sprintf("AI response generated for your question: %s", req.QuestionText)

	doubt := lmsmodels.AiDoubtQuery{
		TenantID:      req.TenantID,
		StudentID:     req.StudentID,
		CourseID:      req.CourseID,
		ReferenceType: req.ReferenceType,
		ReferenceID:   req.ReferenceID,
		QuestionText:  req.QuestionText,
		AiResponse:    aiResponse,
	}

	if err := database.DB.Create(&doubt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save AI doubt"})
		return
	}

	c.JSON(http.StatusOK, doubt)
}

func GetEnrolledCoursesHandler(c *gin.Context) {
	studentId := c.Param("id")
	tenantId := c.Query("tenant_id") // Optional filter

	var courses []lmsmodels.Course
	
	query := database.DB.Table("courses").
		Select("courses.*").
		Joins("JOIN batches ON batches.course_id = courses.id").
		Joins("JOIN enrollments ON enrollments.batch_id = batches.id").
		Where("enrollments.student_id = ?", studentId)

	if tenantId != "" {
		query = query.Where("courses.tenant_id = ?", tenantId)
	}

	if err := query.Find(&courses).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch enrolled courses"})
		return
	}

	c.JSON(http.StatusOK, courses)
}
