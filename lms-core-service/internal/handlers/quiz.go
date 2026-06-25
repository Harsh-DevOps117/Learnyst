package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"lms-core-service/internal/database"
	lmsmodels "lms-core-service/internal/models"
)

type QuizAttemptRequest struct {
	StudentID uint `json:"student_id" binding:"required"`
	Score     *int `json:"score"`
}

func SubmitQuizAttemptHandler(c *gin.Context) {
	quizId := c.Param("id")
	var req QuizAttemptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var qID uint
	fmt.Sscan(quizId, &qID)

	attempt := lmsmodels.QuizAttempt{
		QuizID:    qID,
		StudentID: req.StudentID,
		Score:     req.Score,
		Status:    "completed",
	}

	if err := database.DB.Create(&attempt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit quiz attempt"})
		return
	}

	c.JSON(http.StatusOK, attempt)
}

func GetQuizSummaryHandler(c *gin.Context) {
	quizId := c.Param("id")

	// Mocking AI summary logic
	summary := fmt.Sprintf("AI Test Summary for Quiz %s: Most students struggled with topic X. You scored in the top 20%%.", quizId)

	c.JSON(http.StatusOK, gin.H{
		"quiz_id":    quizId,
		"ai_summary": summary,
		"avg_score":  75.5,
	})
}
