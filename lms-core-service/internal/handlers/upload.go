package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"lms-core-service/internal/storage"
	"net/http"
	"path/filepath"
	"time"
)

func UploadAssetHandler(c *gin.Context) {
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
	ext := filepath.Ext(header.Filename)
	filename := header.Filename[:len(header.Filename)-len(ext)]
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

	if ext == ".mp4" || ext == ".mov" || ext == ".mkv" {
		localFilename := fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), filename, ext)
		localPath := filepath.Join("/shared", localFilename)
		if err := c.SaveUploadedFile(header, localPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save video locally: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "File uploaded successfully",
			"file_name": header.Filename,
			"url":       "local://" + localPath,
		})
		return
	}

	blobName := fmt.Sprintf("%s%d-%s%s", folderPrefix, time.Now().UnixNano(), filename, ext)
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	publicUrl, err := storage.UploadStream(c.Request.Context(), blobName, file, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to local storage: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "File uploaded successfully",
		"file_name": header.Filename,
		"url":       publicUrl,
	})
}
