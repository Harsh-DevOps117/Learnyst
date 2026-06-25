package router

import (
	"io"
	"net/http"
	"strings"

	"tenant-proxy-service/internal/db"
	"tenant-proxy-service/internal/azureblob"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.NoRoute(ProxyHandler)

	return r
}

func ProxyHandler(c *gin.Context) {
	host := c.Request.Host
	// Remove port if present (e.g., localhost:8080 -> localhost)
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	// 1. Identify Tenant
	// If they access mathacademy.learnyst.com, host = mathacademy.learnyst.com
	// For local testing, we might extract the subdomain if the host ends with a base domain.
	// But let's rely on DB lookup directly or strip the base domain.
	baseDomain := "learnyst.com" // This should ideally come from env
	var lookupKey string

	if strings.HasSuffix(host, "."+baseDomain) {
		// Subdomain approach
		lookupKey = strings.TrimSuffix(host, "."+baseDomain)
	} else {
		// Custom domain approach
		lookupKey = host
	}

	tenant, err := db.GetTenantByHost(lookupKey)
	if err != nil || tenant == nil {
		c.String(http.StatusNotFound, "404 - Academy Not Found or Inactive")
		return
	}

	// 2. Map path to Azure object
	path := c.Request.URL.Path
	// Strip leading slash
	path = strings.TrimPrefix(path, "/")

	// 3. Fetch from Azure Blob
	body, contentType, err := azureblob.FetchFile(c.Request.Context(), tenant.AcademySlug, path)
	if err != nil {
		// Fallback to index.html for SPA routing (React/Vue/Next.js)
		body, contentType, err = azureblob.FetchFile(c.Request.Context(), tenant.AcademySlug, "index.html")
		if err != nil {
			c.String(http.StatusNotFound, "404 - Page Not Found")
			return
		}
	}
	defer body.Close()

	// 4. Stream response back to client
	if contentType != nil {
		c.Writer.Header().Set("Content-Type", *contentType)
	}
	
	// Stream the data from Azure directly to the Gin response writer
	io.Copy(c.Writer, body)
}
