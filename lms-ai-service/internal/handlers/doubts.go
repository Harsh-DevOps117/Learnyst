package handlers

import (
	"net/http"

	"lms-ai-service/internal/ai"
	"lms-ai-service/internal/db"

	"github.com/gin-gonic/gin"
)

type DoubtRequest struct {
	StudentID string `json:"student_id" binding:"required"`
	VideoID   uint   `json:"video_id" binding:"required"`
	Question  string `json:"question" binding:"required"`
}

func AskDoubtHandler(c *gin.Context) {
	var req DoubtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ask Groq
	answer, err := ai.AskDoubt(req.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI processing failed", "details": err.Error()})
		return
	}

	// Save to DB
	doubt := db.AIDoubt{
		StudentID: req.StudentID,
		VideoID:   req.VideoID,
		Question:  req.Question,
		Answer:    answer,
	}

	if err := db.DB.Create(&doubt).Error; err != nil {
		// Even if DB fails, we should still return the AI answer to the student
		c.JSON(http.StatusOK, gin.H{
			"answer":  answer,
			"warning": "Failed to save doubt to database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":     doubt.ID,
		"answer": answer,
	})
}
