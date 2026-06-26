package handlers
import (
	"lms-core-service/internal/database"
	lmsmodels "lms-core-service/internal/models"
	"net/http"
	"github.com/gin-gonic/gin"
)
func GetGlobalUsersHandler(c *gin.Context) {
	var users []lmsmodels.User
	if err := database.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	var safeUsers []map[string]interface{}
	for _, u := range users {
		roleName := "Tenant"
		if u.RoleID == 2 {
			roleName = "Teacher"
		} else if u.RoleID == 3 {
			roleName = "Student"
		} else if u.RoleID == 4 {
			roleName = "SuperAdmin"
		}
		safeUsers = append(safeUsers, map[string]interface{}{
			"id":            u.ID,
			"first_name":    u.FirstName,
			"last_name":     u.LastName,
			"email":         u.Email,
			"phone":         u.Phone,
			"role_id":       u.RoleID,
			"role_name":     roleName,
			"is_active":     u.IsActive,
			"created_at":    u.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, safeUsers)
}
func GetGlobalAnalyticsHandler(c *gin.Context) {
	var totalUsers int64
	if err := database.DB.Model(&lmsmodels.User{}).Count(&totalUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count users"})
		return
	}
	var totalRevenue float64
	type EnrollmentResult struct {
		Price float64
	}
	var results []EnrollmentResult
	if err := database.DB.Table("enrollments").
		Select("courses.price").
		Joins("JOIN batches ON enrollments.batch_id = batches.id").
		Joins("JOIN courses ON batches.course_id = courses.id").
		Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate revenue"})
		return
	}
	for _, r := range results {
		totalRevenue += r.Price
	}
	totalProfit := totalRevenue * 0.20
	c.JSON(http.StatusOK, gin.H{
		"total_users":  totalUsers,
		"total_profit": totalProfit,
		"currency":     "INR",
	})
}
func DeleteUserHandler(c *gin.Context) {
	userId := c.Param("id")
	var user lmsmodels.User
	if err := database.DB.First(&user, userId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if user.RoleID == 4 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete another Super Admin"})
		return
	}
	if err := database.DB.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
