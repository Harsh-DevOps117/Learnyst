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
	"tenant-proxy-service/internal/azureblob"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	_ = godotenv.Load()

	// Initialize Database and S3
	db.InitDB()
	azureblob.InitAzureBlob()

	r := router.SetupRouter()

	// Setup Autocert Manager
	// This automatically provisions Let's Encrypt certificates for requested domains.
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: customHostPolicy,
		Cache:      autocert.DirCache("certs-cache"),
	}

	env := os.Getenv("ENV")
	if env == "production" {
		// In production, we run on ports 80 and 443 with Autocert
		server := &http.Server{
			Addr:    ":443",
			Handler: r,
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
			},
		}

		// Also start HTTP server to redirect to HTTPS and answer ACME HTTP-01 challenges
		go func() {
			log.Println("Starting HTTP server on :80 for ACME challenges and redirects...")
			log.Fatal(http.ListenAndServe(":80", certManager.HTTPHandler(nil)))
		}()

		log.Println("Starting HTTPS Proxy Server on :443...")
		log.Fatal(server.ListenAndServeTLS("", ""))
	} else {
		// Local development
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

// customHostPolicy controls which domains we are willing to request certificates for.
func customHostPolicy(ctx context.Context, host string) error {
	// Look up the host in the DB. If it exists as a custom domain or valid subdomain, allow it.
	tenant, err := db.GetTenantByHost(host)
	if err != nil || tenant == nil {
		return fmt.Errorf("acme/autocert: host not configured: %s", host)
	}
	return nil
}
