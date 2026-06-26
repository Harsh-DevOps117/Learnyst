package handlers
import (
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
	"lms-core-service/internal/database"
	lmsmodels "lms-core-service/internal/models"
	"lms-core-service/internal/payment"
)
type CreateOrderRequest struct {
	CourseID   uint   `json:"course_id" binding:"required"`
	CouponCode string `json:"coupon_code"`
}
type VerifyPaymentRequest struct {
	RazorpayPaymentID   string `json:"razorpay_payment_id" binding:"required"`
	RazorpayOrderID     string `json:"razorpay_order_id" binding:"required"`
	RazorpaySignature   string `json:"razorpay_signature" binding:"required"`
	CourseID            uint   `json:"course_id"`
	StudentID           uint   `json:"student_id"`
}
func CreateOrderHandler(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var course lmsmodels.Course
	if err := database.DB.Preload("Tenant").First(&course, req.CourseID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Course not found"})
		return
	}
	baseAmount := int(course.Price * 100)
	finalAmount := baseAmount
	if req.CouponCode != "" {
		var coupon lmsmodels.Coupon
		// Find active coupon for this tenant and (either this course or all courses)
		if err := database.DB.Where("code = ? AND tenant_id = ? AND is_active = ?", req.CouponCode, course.TenantID, true).
			Where("course_id IS NULL OR course_id = ?", course.ID).
			First(&coupon).Error; err == nil {
			// Apply discount
			discount := (baseAmount * coupon.DiscountPercent) / 100
			// Cap discount at MaxDiscountAmount if set
			maxDiscountPaise := int(coupon.MaxDiscountAmount * 100)
			if maxDiscountPaise > 0 && discount > maxDiscountPaise {
				discount = maxDiscountPaise
			}
			finalAmount = baseAmount - discount
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired coupon code"})
			return
		}
	}
	if finalAmount < 100 {
		finalAmount = 100 
	}
	receipt := fmt.Sprintf("rcpt_crs_%d", course.ID)
	order, err := payment.CreateOrderWithSplit(finalAmount, receipt, course.Tenant.RazorpayAccountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order with Razorpay: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}
func VerifyPaymentHandler(c *gin.Context) {
	var req VerifyPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload: " + err.Error()})
		return
	}
	isValid := payment.VerifySignature(req.RazorpayOrderID, req.RazorpayPaymentID, req.RazorpaySignature)
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signature. Payment verification failed."})
		return
	}
	if req.CourseID > 0 && req.StudentID > 0 {
		var course lmsmodels.Course
		if err := database.DB.First(&course, req.CourseID).Error; err == nil {
			var batch lmsmodels.Batch
			result := database.DB.Where("course_id = ? AND tenant_id = ?", course.ID, course.TenantID).First(&batch)
			if result.Error != nil {
				batch = lmsmodels.Batch{
					TenantID: course.TenantID,
					CourseID: course.ID,
					Name:     "Default Cohort",
					IsActive: true,
				}
				database.DB.Create(&batch)
			}
			var existingEnrollment lmsmodels.Enrollment
			err := database.DB.Where("batch_id = ? AND student_id = ?", batch.ID, req.StudentID).First(&existingEnrollment).Error
			if err != nil { 
				enrollment := lmsmodels.Enrollment{
					TenantID:  course.TenantID,
					BatchID:   batch.ID,
					StudentID: req.StudentID,
					Status:    "active",
				}
				if err := database.DB.Create(&enrollment).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enroll student"})
					return
				}
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Payment verified and student enrolled successfully",
		"status":  "paid",
	})
}
type UpdatePaymentDetailsRequest struct {
	UpiID             string `json:"upi_id"`
	RazorpayAccountID string `json:"razorpay_account_id"`
}
func UpdatePaymentDetailsHandler(c *gin.Context) {
	var req UpdatePaymentDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userId := c.Param("id")
	var tenant lmsmodels.Tenant
	if err := database.DB.Where("owner_user_id = ?", userId).First(&tenant).Error; err == nil {
		tenant.UpiID = req.UpiID
		tenant.RazorpayAccountID = req.RazorpayAccountID
		database.DB.Save(&tenant)
	}
	database.DB.Model(&lmsmodels.TenantStaff{}).
		Where("user_id = ?", userId).
		Update("upi_id", req.UpiID)
	c.JSON(http.StatusOK, gin.H{"message": "Payment details updated successfully"})
}
func GetTenantPayoutsHandler(c *gin.Context) {
	tenantId := c.Param("tenantId")
	var totalRevenue float64
	type EnrollmentResult struct {
		Price float64
	}
	var results []EnrollmentResult
	database.DB.Table("enrollments").
		Select("courses.price").
		Joins("JOIN batches ON enrollments.batch_id = batches.id").
		Joins("JOIN courses ON batches.course_id = courses.id").
		Where("enrollments.tenant_id = ?", tenantId).
		Scan(&results)
	for _, r := range results {
		totalRevenue += r.Price
	}
	platformFee := totalRevenue * 0.20
	tenantNet := totalRevenue * 0.80
	c.JSON(http.StatusOK, gin.H{
		"tenant_id":        tenantId,
		"total_revenue":    totalRevenue,
		"platform_fee":     platformFee,
		"tenant_net":       tenantNet,
		"pending_payouts":  tenantNet * 0.1, 
		"currency":         "INR",
		"split_ratio":      "80_20_tenant_platform",
	})
}
