package router
import (
	"io"
	"net"
	"net/http"
	"os"
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
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	baseDomain := os.Getenv("BASE_DOMAIN")
	if baseDomain == "" {
		baseDomain = "localhost" 
	}
	var lookupKey string
	if strings.HasSuffix(host, "."+baseDomain) {
		lookupKey = strings.TrimSuffix(host, "."+baseDomain)
	} else if host == baseDomain {
		lookupKey = "main"
	} else {
		lookupKey = host
		if baseDomain != "localhost" {
			txtRecords, err := net.LookupTXT(lookupKey)
			if err != nil {
				c.String(http.StatusForbidden, "403 - DNS Validation Failed")
				return
			}
			valid := false
			for _, txt := range txtRecords {
				if strings.Contains(txt, "learnyst-verification=") {
					valid = true
					break
				}
			}
			if !valid {
				c.String(http.StatusForbidden, "403 - Domain Not Verified")
				return
			}
		}
	}
	tenant, err := db.GetTenantByHost(lookupKey)
	if err != nil || tenant == nil {
		c.String(http.StatusNotFound, "404 - Academy Not Found or Inactive")
		return
	}
	path := c.Request.URL.Path
	path = strings.TrimPrefix(path, "/")
	body, contentType, err := azureblob.FetchFile(c.Request.Context(), tenant.AcademySlug, path)
	if err != nil {
		body, contentType, err = azureblob.FetchFile(c.Request.Context(), tenant.AcademySlug, "index.html")
		if err != nil {
			c.String(http.StatusNotFound, "404 - Page Not Found")
			return
		}
	}
	defer body.Close()
	if contentType != nil {
		c.Writer.Header().Set("Content-Type", *contentType)
	}
	io.Copy(c.Writer, body)
}
