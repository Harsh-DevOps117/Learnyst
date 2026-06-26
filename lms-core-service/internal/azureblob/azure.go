package azureblob
import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
)
var Client *azblob.Client
var AssetsContainerName string
var AccountName string
func InitAzureBlob() {
	AccountName = os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	AssetsContainerName = os.Getenv("AZURE_ASSETS_CONTAINER")
	if AccountName == "" || accountKey == "" {
		log.Println("WARNING: Azure Storage credentials not set in Core Service")
		return
	}
	if AssetsContainerName == "" {
		AssetsContainerName = "learnyst-tenant-assets" 
	}
	cred, err := azblob.NewSharedKeyCredential(AccountName, accountKey)
	if err != nil {
		log.Fatalf("Invalid Azure credentials: %v", err)
	}
	blobURL := fmt.Sprintf("https://%s.blob.core.windows.net/", AccountName)
	Client, err = azblob.NewClientWithSharedKeyCredential(blobURL, cred, nil)
	if err != nil {
		log.Fatalf("Unable to create Azure Blob Client, %v", err)
	}
	_, _ = Client.CreateContainer(context.Background(), AssetsContainerName, nil)
	log.Println("Azure Blob Client initialized successfully in Core Service")
}
func UploadStream(ctx context.Context, blobName string, fileStream io.Reader, contentType string) (string, error) {
	if Client == nil {
		return "", fmt.Errorf("Azure Blob Client not initialized")
	}
	opts := &azblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: &contentType,
		},
	}
	_, err := Client.UploadStream(ctx, AssetsContainerName, blobName, fileStream, opts)
	if err != nil {
		return "", fmt.Errorf("failed to upload blob: %v", err)
	}
	publicUrl := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", AccountName, AssetsContainerName, blobName)
	return publicUrl, nil
}
func DeleteBlob(ctx context.Context, blobUrlOrName string) error {
	if Client == nil {
		return fmt.Errorf("Azure Blob Client not initialized")
	}
	prefix := fmt.Sprintf("https://%s.blob.core.windows.net/%s/", AccountName, AssetsContainerName)
	blobName := blobUrlOrName
	if len(blobUrlOrName) > len(prefix) && blobUrlOrName[:len(prefix)] == prefix {
		blobName = blobUrlOrName[len(prefix):]
	}
	_, err := Client.DeleteBlob(ctx, AssetsContainerName, blobName, nil)
	return err
}
