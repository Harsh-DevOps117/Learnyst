package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"lms-core-service/internal/azureblob"
)

// UploadAssetHandler handles generic file uploads for logos, banners, and assignment PDFs.
func UploadAssetHandler(c *gin.Context) {
	// Parse the multipart form
	// Limit upload size to 10 MB (can be adjusted)
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeds limit or invalid form"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded. Expected 'file' field."})
		return
	}
	defer file.Close()

	// You could extract tenant ID or user ID from JWT to organize folders
	// e.g. path := fmt.Sprintf("tenant_%d/%s", tenantID, newFileName)
	// For now, we'll just put them in the root of the assets container with a timestamp.
	
	ext := filepath.Ext(header.Filename)
	filename := header.Filename[:len(header.Filename)-len(ext)]
	
	// Read slug, phone, and category from form
	slug := c.PostForm("slug")
	phone := c.PostForm("phone")
	category := c.PostForm("category")

	folderPrefix := ""
	if slug != "" && phone != "" {
		folderPrefix = fmt.Sprintf("%s_%s/", slug, phone)
	} else if slug != "" {
		folderPrefix = fmt.Sprintf("%s/", slug)
	}
	
	if category != "" {
		folderPrefix = folderPrefix + category + "/"
	}

	// Create a unique blob name to prevent overwrites
	blobName := fmt.Sprintf("%s%d-%s%s", folderPrefix, time.Now().UnixNano(), filename, ext)

	// Determine Content Type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload to Azure
	publicUrl, err := azureblob.UploadStream(c.Request.Context(), blobName, file, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to Azure: " + err.Error()})
		return
	}

	// Return the public URL
	c.JSON(http.StatusOK, gin.H{
		"message":   "File uploaded successfully",
		"file_name": header.Filename,
		"url":       publicUrl,
	})
}
