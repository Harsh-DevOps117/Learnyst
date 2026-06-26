package azureblob
import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)
func DeleteBlobFromUrl(blobUrl string) error {
	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	containerName := os.Getenv("AZURE_ASSETS_CONTAINER")
	if accountName == "" || accountKey == "" || containerName == "" {
		return fmt.Errorf("azure storage credentials not configured")
	}
	parsedUrl, err := url.Parse(blobUrl)
	if err != nil {
		return fmt.Errorf("failed to parse blob URL: %w", err)
	}
	pathParts := strings.SplitN(parsedUrl.Path, "/", 3)
	if len(pathParts) < 3 || pathParts[1] != containerName {
		return fmt.Errorf("invalid blob URL or container mismatch: %s", blobUrl)
	}
	blobName := pathParts[2] 
	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return fmt.Errorf("failed to create azure credential: %w", err)
	}
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create azure client: %w", err)
	}
	_, err = client.DeleteBlob(context.TODO(), containerName, blobName, nil)
	if err != nil {
		return fmt.Errorf("failed to delete blob %s: %w", blobName, err)
	}
	return nil
}
