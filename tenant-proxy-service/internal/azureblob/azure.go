package azureblob

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

var Client *azblob.Client
var ContainerName string

func InitAzureBlob() {
	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	ContainerName = os.Getenv("AZURE_CONTAINER_NAME")

	if accountName == "" || accountKey == "" {
		log.Println("WARNING: Azure Storage credentials not set, relying on defaults")
		return
	}

	if ContainerName == "" {
		ContainerName = "learnyst-tenant-sites" // Fallback
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

func FetchFile(ctx context.Context, tenantSlug, path string) (io.ReadCloser, *string, error) {
	if path == "" || path == "/" {
		path = "index.html"
	}

	blobName := tenantSlug + "/" + path

	if Client == nil {
		return nil, nil, fmt.Errorf("Azure Blob Client not initialized")
	}

	resp, err := Client.DownloadStream(ctx, ContainerName, blobName, nil)
	if err != nil {
		return nil, nil, err
	}

	return resp.Body, resp.ContentType, nil
}
