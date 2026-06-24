package main

import (
	"lms-api-gateway/internal/config"
	"lms-api-gateway/internal/db"
	"lms-api-gateway/internal/service"

	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)


func main(){

	if err := godotenv.Load(".env"); err != nil { config.Log.Fatal("Failed to load .env file") }

	PORT:=os.Getenv("PORT")
	ENV:=os.Getenv("ENV")

	if PORT=="" { PORT="3000"}
	if ENV=="" { ENV="development"}

	config.Log.Info("Setting Up PORT and ENVIRONMENT")

	db, err:=db.Initdatabase()
	if err != nil {
		config.Log.Fatal(err.Error())
	}
	db.AutoMigrate(db)

	config.InitLogger()
	config.Log.Info("Initialising The Server....")

	gin.SetMode(gin.DebugMode)

	app:=gin.New()
	app.GET("/health",service.NewHealthCheckHandler().HealthCheck)

	v1Auth := app.Group("/v1/auth")
	{
		v1Auth.POST("/register", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"status":  http.StatusOK,
				"time":     time.Now().UnixMilli(),
				"message": "Register placeholder route running asynchronously"})
		})
		v1Auth.POST("/login", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"status":  http.StatusOK,
				"time":     time.Now().UnixMilli(),
				"message": "Login placeholder route running asynchronously"})
		})
	}
	config.Log.Info("🟢 Auth Microservice running flawlessly", zap.String("port", PORT), zap.String("env", ENV))
	app.Run(":"+PORT)
}
