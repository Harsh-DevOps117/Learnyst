package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
)

func main() {
	AccountName := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")

	if AccountName == "" || accountKey == "" {
		log.Fatalf("Credentials not set")
	}

	cred, err := azblob.NewSharedKeyCredential(AccountName, accountKey)
	if err != nil {
		log.Fatalf("Invalid Azure credentials: %v", err)
	}

	blobURL := fmt.Sprintf("https://%s.blob.core.windows.net/", AccountName)
	client, err := azblob.NewClientWithSharedKeyCredential(blobURL, cred, nil)
	if err != nil {
		log.Fatalf("Unable to create Azure Blob Client: %v", err)
	}

	allowedOrigins := "*"
	allowedMethods := "GET,OPTIONS,PUT,POST,HEAD"
	allowedHeaders := "*"
	exposedHeaders := "*"
	maxAge := int32(86400)

	corsRules := []*service.CORSRule{
		{
			AllowedOrigins:  &allowedOrigins,
			AllowedMethods:  &allowedMethods,
			AllowedHeaders:  &allowedHeaders,
			ExposedHeaders:  &exposedHeaders,
			MaxAgeInSeconds: &maxAge,
		},
	}

	_, err = client.ServiceClient().SetProperties(context.Background(), &service.SetPropertiesOptions{
		CORS: corsRules,
	})
	if err != nil {
		log.Fatalf("Failed to set CORS properties: %v", err)
	}

	fmt.Println("Successfully configured Azure Blob Storage CORS to allow all origins.")
}
