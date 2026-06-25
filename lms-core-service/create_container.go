package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	containerName := os.Getenv("AZURE_ASSETS_CONTAINER")

	if containerName == "" {
		containerName = "learnyst-tenant-assets"
	}

	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Fatalf("Invalid credentials: %v", err)
	}

	blobURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
	client, err := azblob.NewClientWithSharedKeyCredential(blobURL, cred, nil)
	if err != nil {
		log.Fatalf("Unable to create client: %v", err)
	}

	access := azblob.PublicAccessTypeBlob
	opts := &azblob.CreateContainerOptions{
		Access: &access,
	}

	fmt.Printf("Creating container '%s' in account '%s'...\n", containerName, accountName)
	_, err = client.CreateContainer(context.Background(), containerName, opts)
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}

	fmt.Println("Container created successfully with Blob public access!")
}
