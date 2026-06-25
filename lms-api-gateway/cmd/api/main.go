package main

import (
	"lms-api-gateway/internal/config"
	"lms-api-gateway/internal/controller"
	database "lms-api-gateway/internal/database"
	"lms-api-gateway/internal/service"
	"lms-api-gateway/pkg/middleware"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	config.InitLogger()

	dir, _ := os.Getwd()
	config.Log.Info("Current working directory is: ", zap.String("dir", dir))

	if err := godotenv.Load(".env"); err != nil {
		config.Log.Warn("Failed to load .env file")
	}

	PORT := os.Getenv("PORT")
	ENV := os.Getenv("ENV")
	REDIS_URI := os.Getenv("REDIS_URI")

	if PORT == "" {
		PORT = "3000"
	}
	if ENV == "" {
		ENV = "development"
	}
	if REDIS_URI == "" {
		REDIS_URI = os.Getenv("REDIS_URL") // Fallback
		if REDIS_URI == "" {
			REDIS_URI = "redis://localhost:6379"
		}
	}

	config.Log.Info("Setting Up PORT and ENVIRONMENT")

	db, err := database.Initdatabase()
	if err != nil {
		config.Log.Fatal(err.Error())
	}

	database.Automigrate(db)
	config.Log.Info("Initialising The Server....")

	// Init Redis
	redisOptions, err := redis.ParseURL(REDIS_URI)
	if err != nil {
		config.Log.Fatal("Failed to parse REDIS_URL: " + err.Error())
	}
	redisClient := redis.NewClient(redisOptions)
	redisRepo := config.NewRedisRepo(redisClient)

	// Init Services
	authService := service.NewAuthService(db, redisRepo)

	// Init Controllers
	authController := controller.NewAuthController(authService)

	gin.SetMode(gin.DebugMode)

	app := gin.New()
	
	// Basic CORS for frontend integration
	app.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	app.GET("/health", service.NewHealthCheckHandler().HealthCheck)

	v1Auth := app.Group("/v1/auth")
	{
		// Unprotected Routes
		v1Auth.POST("/register/initiate", authController.InitiateSignup)
		v1Auth.GET("/register/resend-otp", authController.ResendSignupOTP)
		v1Auth.POST("/register/complete", authController.CompleteSignup)
		v1Auth.POST("/login", authController.Login)
		v1Auth.POST("/reset-password", authController.RequestPasswordReset)
		v1Auth.POST("/reset-password/complete", authController.CompletePasswordReset)

		// Protected Routes
		protected := v1Auth.Group("")
		protected.Use(middleware.JWTAuthMiddleware())
		{
			protected.POST("/change-password", authController.ChangePassword)
			protected.POST("/change-email/initiate", authController.InitiateChangeEmail)
			protected.POST("/change-email/verify", authController.VerifyChangeEmail)
		}
	}
	config.Log.Info("🟢 Auth Microservice running flawlessly", zap.String("port", PORT), zap.String("env", ENV))
	app.Run(":" + PORT)
}
