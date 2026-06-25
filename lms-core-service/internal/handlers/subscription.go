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
	PlanID   string `json:"plan_id" binding:"required"` // Razorpay Plan ID (e.g., plan_XXX)
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
		"total_count": 120, // Example: bill for 120 cycles (10 years if monthly)
		"customer_notify": 1,
		"notes": map[string]string{
			"tenant_id": fmt.Sprintf("%d", tenant.ID),
		},
	}

	// NOTE: razorpay-go doesn't have a strongly typed Subscription struct in older versions.
	// We use the generic wrapper to hit the subscriptions endpoint if it exists, or raw HTTP.
	// We'll use the Addon / Subscription endpoint.
	// razorpay.NewClient(...).Subscription.Create(...) is typically available in newer SDKs.
	// We'll use a mocked success if the SDK method fails or isn't present in this specific version,
	// but standard Razorpay Go SDK supports it via client.Subscription.Create(data, nil).
	
	body, err := client.Subscription.Create(data, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription: " + err.Error()})
		return
	}

	// Optionally save the subscription ID to the database or mark as "pending_payment"
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription created successfully",
		"subscription_id": body["id"],
		"short_url": body["short_url"], // Razorpay returns a hosted checkout link for the subscription
	})
}
