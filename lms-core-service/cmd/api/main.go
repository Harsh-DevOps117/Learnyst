package main
import (
	"log"
	"os"
	"lms-core-service/internal/azureblob"
	"lms-core-service/internal/database"
	"lms-core-service/internal/handlers"
	"lms-core-service/internal/kafka"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)
func main() {
	_ = godotenv.Load()
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001" 
	}
	log.Println("Initializing LMS Core Service Database...")
	db, err := database.InitDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	err = database.AutoMigrate(db)
	if err != nil {
		log.Fatal("Failed to auto-migrate: ", err)
	}
	database.InitDBResolver(db)
	log.Println("Initializing Azure Blob Service...")
	azureblob.InitAzureBlob()
	log.Println("Initializing Kafka Producer and Consumer...")
	go kafka.InitProducer()
	go kafka.StartConsumer(context.Background())
	log.Println("Starting LMS Core Service on port", port)
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "lms-core-service"})
	})
	v1 := router.Group("/api/v1")
	{
		v1.POST("/invitations/accept", handlers.AcceptInvitationHandler)
		v1.POST("/upload", handlers.UploadAssetHandler)
		users := v1.Group("/users")
		{
			users.GET("/:id", handlers.GetUserHandler)
		users.PUT("/:id/profile-image", handlers.UpdateUserProfileHandler)
			users.PUT("/:id/payment-details", handlers.UpdatePaymentDetailsHandler)
		}
		tenants := v1.Group("/tenants")
		{
			tenants.POST("", handlers.CreateTenantHandler)
			tenants.GET("/slug/:slug", handlers.GetTenantBySlugHandler)
			tenants.PUT("/:id/branding", handlers.UpdateTenantBrandingHandler)
			tenants.PUT("/:id/domain", handlers.UpdateTenantDomainHandler)
			tenants.GET("/:id/analytics", handlers.GetTenantAnalyticsHandler)
			tenants.GET("/:id/teachers", handlers.GetTenantTeachersHandler)
			tenants.GET("/:id/courses", handlers.GetTenantCoursesHandler)
			tenants.POST("/:tenant_id/invitations", handlers.SendInvitationHandler)
		}
		courses := v1.Group("/courses")
		{
			courses.POST("", handlers.CreateCourseHandler)
			courses.PUT("/:id", handlers.UpdateCourseHandler)
			courses.POST("/:id/videos", handlers.CreateCourseVideoHandler)
			courses.POST("/:id/assignments", handlers.CreateAssignmentHandler)
			courses.GET("/:id/consume", handlers.ConsumeCourseHandler)
			courses.DELETE("/:id", handlers.DeleteCourseHandler)
		}
		progress := v1.Group("/progress")
		{
			progress.POST("/video", handlers.TrackVideoProgressHandler)
		}
		students := v1.Group("/students")
		{
			students.GET("/:id/courses", handlers.GetEnrolledCoursesHandler)
		}
		quizzes := v1.Group("/quizzes")
		{
			quizzes.POST("/:id/attempts", handlers.SubmitQuizAttemptHandler)
			quizzes.GET("/:id/summary", handlers.GetQuizSummaryHandler)
		}
		ai := v1.Group("/ai")
		{
			ai.POST("/doubt", handlers.CreateDoubtQueryHandler)
		}
		payments := v1.Group("/payments")
		{
			payments.POST("/checkout", handlers.CreateOrderHandler)
			payments.POST("/create-order", handlers.CreateOrderHandler) 
			payments.POST("/verify", handlers.VerifyPaymentHandler)
			payments.GET("/payouts/:tenantId", handlers.GetTenantPayoutsHandler)
		}
		admin := v1.Group("/admin")
		{
			admin.GET("/users", handlers.GetGlobalUsersHandler)
			admin.GET("/analytics", handlers.GetGlobalAnalyticsHandler)
			admin.DELETE("/users/:id", handlers.DeleteUserHandler)
		}
		subscriptions := v1.Group("/subscriptions")
		{
			subscriptions.POST("/create", handlers.CreateTenantSubscriptionHandler)
		}
	}
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
