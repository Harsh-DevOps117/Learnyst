package service

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthCheckHandler struct {}

func NewHealthCheckHandler() *HealthCheckHandler {
	return &HealthCheckHandler{}
}

func (h *HealthCheckHandler) HealthCheck(c *gin.Context){
	c.JSON(http.StatusOK,gin.H{
		"status":      "UP",
		"version":     "1.0.0",
		"service":     "lms-auth-service",
		"time":        time.Now(),
		"environment": "development",
	})
}
