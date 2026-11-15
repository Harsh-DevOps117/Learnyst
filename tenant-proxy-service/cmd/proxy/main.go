package main
import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"tenant-proxy-service/internal/db"
	"tenant-proxy-service/internal/router"
	"tenant-proxy-service/internal/storage"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/acme/autocert"
)
func main() {
	_ = godotenv.Load()
	db.InitDB()
	storage.InitStorage()
	r := router.SetupRouter()
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: customHostPolicy,
		Cache:      autocert.DirCache("certs-cache"),
	}
	env := os.Getenv("ENV")
	if env == "production" {
		server := &http.Server{
			Addr:    ":443",
			Handler: r,
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
			},
		}
		go func() {
			log.Println("Starting HTTP server on :80 for ACME challenges and redirects...")
			log.Fatal(http.ListenAndServe(":80", certManager.HTTPHandler(nil)))
		}()
		log.Println("Starting HTTPS Proxy Server on :443...")
		log.Fatal(server.ListenAndServeTLS("", ""))
	} else {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		log.Println("Starting Local Proxy Server on :" + port)
		if err := r.Run(":" + port); err != nil {
			log.Fatal("Failed to start server: ", err)
		}
	}
}
func customHostPolicy(ctx context.Context, host string) error {
	tenant, err := db.GetTenantByHost(host)
	if err != nil || tenant == nil {
		return fmt.Errorf("acme/autocert: host not configured: %s", host)
	}
	return nil
}
