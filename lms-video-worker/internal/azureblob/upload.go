package azureblob
import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)
func UploadFile(localFilePath, blobName string) (string, error) {
	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	containerName := os.Getenv("AZURE_ASSETS_CONTAINER")
	if accountName == "" || accountKey == "" || containerName == "" {
		return "", fmt.Errorf("azure storage credentials not configured")
	}
	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return "", fmt.Errorf("failed to create azure credential: %w", err)
	}
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create azure client: %w", err)
	}
	file, err := os.Open(localFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", localFilePath, err)
	}
	defer file.Close()
	ext := filepath.Ext(localFilePath)
	contentType := "application/octet-stream"
	if ext == ".m3u8" {
		contentType = "application/vnd.apple.mpegurl"
	} else if ext == ".ts" {
		contentType = "video/MP2T"
	}
	_, err = client.UploadFile(context.TODO(), containerName, blobName, file,
		&azblob.UploadFileOptions{
			HTTPHeaders: &blob.HTTPHeaders{
				BlobContentType: &contentType,
			},
		})
	if err != nil {
		return "", fmt.Errorf("failed to upload blob %s: %w", blobName, err)
	}
	blobURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", accountName, containerName, blobName)
	return blobURL, nil
}
