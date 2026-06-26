package handlers
import (
	"fmt"
	"net/http"
	"os"
	"github.com/gin-gonic/gin"
	"github.com/razorpay/razorpay-go"
	"lms-core-service/internal/database"
	lmsmodels "lms-core-service/internal/models"
)
type CreateSubscriptionRequest struct {
	TenantID uint `json:"tenant_id" binding:"required"`
	PlanID   string `json:"plan_id" binding:"required"` 
}
func CreateTenantSubscriptionHandler(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var tenant lmsmodels.Tenant
	if err := database.DB.First(&tenant, req.TenantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}
	keyId := os.Getenv("RAZORPAY_KEY_ID")
	keySecret := os.Getenv("RAZORPAY_KEY_SECRET")
	client := razorpay.NewClient(keyId, keySecret)
	data := map[string]interface{}{
		"plan_id":     req.PlanID,
		"total_count": 120, 
		"customer_notify": 1,
		"notes": map[string]string{
			"tenant_id": fmt.Sprintf("%d", tenant.ID),
		},
	}
	body, err := client.Subscription.Create(data, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription created successfully",
		"subscription_id": body["id"],
		"short_url": body["short_url"], 
	})
}
