package handlers
import (
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
	"lms-core-service/internal/database"
	lmsmodels "lms-core-service/internal/models"
)
type CreateTenantRequest struct {
	OwnerUserID uint   `json:"owner_user_id" binding:"required"`
	AcademyName string `json:"academy_name" binding:"required"`
	AcademySlug string `json:"academy_slug" binding:"required"`
}
func CreateTenantHandler(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tenant := lmsmodels.Tenant{
		OwnerUserID: req.OwnerUserID,
		AcademyName: req.AcademyName,
		AcademySlug: req.AcademySlug,
	}
	if err := database.DB.Create(&tenant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
		return
	}
	c.JSON(http.StatusOK, tenant)
}
type UpdateBrandingRequest struct {
	LogoUrl   string `json:"logo_url"`
	BannerUrl string `json:"banner_url"`
}
func UpdateTenantBrandingHandler(c *gin.Context) {
	id := c.Param("id")
	var req UpdateBrandingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updates := map[string]interface{}{}
	if req.LogoUrl != "" {
		updates["logo_url"] = req.LogoUrl
	}
	if req.BannerUrl != "" {
		updates["banner_url"] = req.BannerUrl
	}
	if len(updates) > 0 {
		if err := database.DB.Model(&lmsmodels.Tenant{}).Where("id = ?", id).
			Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update branding"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Branding updated successfully"})
}
type UpdateDomainRequest struct {
	CustomDomain string `json:"custom_domain"`
	SSLEnabled   bool   `json:"ssl_enabled"`
}
func UpdateTenantDomainHandler(c *gin.Context) {
	id := c.Param("id")
	var req UpdateDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := database.DB.Model(&lmsmodels.Tenant{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"custom_domain": req.CustomDomain,
			"ssl_enabled":   req.SSLEnabled,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update domain"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Domain updated successfully"})
}
func GetTenantAnalyticsHandler(c *gin.Context) {
	id := c.Param("id")
	var totalStudents int64
	database.DB.Model(&lmsmodels.Enrollment{}).Where("tenant_id = ?", id).Count(&totalStudents)
	var totalCourses int64
	database.DB.Model(&lmsmodels.Course{}).Where("tenant_id = ?", id).Count(&totalCourses)
	var totalRevenue float64
	type EnrollmentResult struct {
		Price float64
	}
	var results []EnrollmentResult
	database.DB.Table("enrollments").
		Select("courses.price").
		Joins("JOIN batches ON enrollments.batch_id = batches.id").
		Joins("JOIN courses ON batches.course_id = courses.id").
		Where("enrollments.tenant_id = ?", id).
		Scan(&results)
	for _, r := range results {
		totalRevenue += r.Price
	}
	c.JSON(http.StatusOK, gin.H{
		"tenant_id":      id,
		"total_students": totalStudents,
		"total_courses":  totalCourses,
		"total_revenue":  totalRevenue,
		"ai_summary":     "The academy is performing well. Advanced React Patterns is your top selling course.",
	})
}
func GetTenantTeachersHandler(c *gin.Context) {
	id := c.Param("id")
	var teachers []lmsmodels.TenantStaff
	if err := database.DB.Preload("User").
		Where("tenant_id = ? AND role_id = ?", id, 2). 
		Find(&teachers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teachers"})
		return
	}
	var response []map[string]interface{}
	for _, t := range teachers {
		response = append(response, map[string]interface{}{
			"staff_id": t.ID,
			"user_id":  t.UserID,
			"name":     t.User.FirstName + " " + t.User.LastName,
			"email":    t.User.Email,
			"upi_id":   t.UpiID,
		})
	}
	c.JSON(http.StatusOK, response)
}
func GetTenantCoursesHandler(c *gin.Context) {
	id := c.Param("id")
	var courses []lmsmodels.Course
	if err := database.DB.Preload("Videos").Preload("Assignments").Where("tenant_id = ?", id).Find(&courses).Error; err != nil {
		log.Printf("DB Error fetching courses for tenant %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch courses: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, courses)
}
func GetTenantBySlugHandler(c *gin.Context) {
	slug := c.Param("slug")
	var tenant lmsmodels.Tenant
	if err := database.DB.Preload("Owner").
		Where("academy_slug = ?", slug).
		First(&tenant).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
		return
	}
	var courses []lmsmodels.Course
	database.DB.Where("tenant_id = ?", tenant.ID).Find(&courses)
	c.JSON(http.StatusOK, gin.H{
		"tenant": tenant,
		"courses": courses,
	})
}
