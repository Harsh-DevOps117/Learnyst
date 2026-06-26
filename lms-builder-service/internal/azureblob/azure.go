package azureblob
import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)
var Client *azblob.Client
var ContainerName string
func InitAzureBlob() {
	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	ContainerName = os.Getenv("AZURE_CONTAINER_NAME")
	if accountName == "" || accountKey == "" {
		log.Println("WARNING: Azure Storage credentials not set")
		return
	}
	if ContainerName == "" {
		ContainerName = "learnyst-tenant-sites"
	}
	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Fatalf("Invalid Azure credentials: %v", err)
	}
	blobURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
	Client, err = azblob.NewClientWithSharedKeyCredential(blobURL, cred, nil)
	if err != nil {
		log.Fatalf("Unable to create Azure Blob Client, %v", err)
	}
	log.Println("Azure Blob Client initialized successfully")
}
func UploadDirectory(tenantSlug string, localDir string) error {
	log.Printf("Starting Azure Blob upload for %s from %s", tenantSlug, localDir)
	if Client == nil {
		log.Println("Mock Upload (Azure Client not initialized)")
		return nil
	}
	err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(localDir, path)
		blobName := tenantSlug + "/" + relPath
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		log.Printf("Mock Uploaded: azure://%s/%s", ContainerName, blobName)
		return nil
	})
	return err
}
