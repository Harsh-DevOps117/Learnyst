package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"lms-api-gateway/internal/controller"
	database "lms-api-gateway/internal/database"
	"lms-api-gateway/internal/service"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {
	_ = godotenv.Load()
	db, err := database.Initdatabase()
	if err != nil {
		log.Fatal("Auth database connection failed")
	}
	if err := database.Automigrate(db); err != nil {
		log.Fatal("Auth session migration failed: ", err)
	}
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.Use(gin.Recovery())
	// Never trust arbitrary X-Forwarded-For headers for rate-limit identity.
	_ = app.SetTrustedProxies(nil)
	app.Use(func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			origin := c.GetHeader("Origin")
			allowed := strings.TrimRight(os.Getenv("FRONTEND_URL"), "/")
			if origin != "" && origin != allowed {
				c.AbortWithStatusJSON(403, gin.H{"error": "Request origin is not allowed."})
				return
			}
		}
		c.Next()
	})
	app.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok", "service": "auth"}) })
	ac := controller.NewAuthController(service.NewAuthService(db))
	var mu sync.Mutex
	type bucket struct {
		Count int
		Reset time.Time
	}
	limits := map[string]bucket{}
	throttle := func(c *gin.Context) {
		mu.Lock()
		now := time.Now()
		for key, b := range limits {
			if now.After(b.Reset) {
				delete(limits, key)
			}
		}
		key := c.ClientIP() + ":" + c.FullPath()
		b := limits[key]
		if b.Reset.IsZero() {
			b.Reset = now.Add(15 * time.Minute)
		}
		b.Count++
		limits[key] = b
		mu.Unlock()
		if b.Count > 30 {
			c.Header("Retry-After", "900")
			c.AbortWithStatusJSON(429, gin.H{"error": "Too many attempts. Please try again in 15 minutes."})
			return
		}
		c.Next()
	}
	auth := app.Group("/v1/auth")
	auth.GET("/config", func(c *gin.Context) {
		c.JSON(200, gin.H{"demo_enabled": os.Getenv("SEED_DEMO") == "true" && os.Getenv("ENV") == "development", "local_inbox_url": "http://localhost:8025"})
	})
	auth.GET("/session", ac.Session)
	auth.POST("/logout", ac.Logout)
	auth.POST("/login", throttle, ac.Login)
	auth.POST("/register", throttle, ac.Register)
	auth.POST("/reset-password", throttle, ac.RequestPasswordReset)
	auth.POST("/reset-password/complete", throttle, ac.CompletePasswordReset)
	auth.POST("/change-password", throttle, ac.ChangePassword)
	port := os.Getenv("PORT")
	if port == "" {
		port = "3010"
	}
	log.Print("Learnyst auth listening on http://127.0.0.1:", port)
	if err := app.Run("127.0.0.1:" + port); err != nil {
		log.Fatal(err)
	}
}
