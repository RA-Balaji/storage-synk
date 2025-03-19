package gcp

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func HMACKeyCreate(ctx context.Context, serviceAccountEmail, projectID string) (storage.HMACKey, error) {
	// Create a new Storage client.
	client, err := storage.NewClient(ctx)
	if err != nil {
		return storage.HMACKey{}, fmt.Errorf("failed to create storage client: %v", err)
	}
	defer client.Close()

	// Generate the HMAC key.
	key, err := client.CreateHMACKey(ctx, projectID, serviceAccountEmail)
	if err != nil {
		return storage.HMACKey{}, fmt.Errorf("Failed to create HMAC key: %v", err)
	}

	fmt.Printf("HMAC Key created successfully:\n")
	fmt.Printf("Access ID: %s\n", key.AccessID)
	fmt.Printf("Secret: %s\n", key.Secret)

	return *key, nil
}

func GcsDownload(ctx context.Context, bucketName, destinationPath string) error {

	localFolder := filepath.Join(destinationPath, bucketName)
	if _, err := os.Stat(localFolder); os.IsNotExist(err) {
		err := os.Mkdir(localFolder, os.ModeDir)
		if err != nil {
			return fmt.Errorf(
				"Error creating directory [%s] at [%s] Err:[%v]", bucketName, destinationPath, err)
		}
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %v", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)

	var wg sync.WaitGroup

	it := bucket.Objects(ctx, nil)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("Error iterating Objects: %v", err)
		}

		wg.Add(1)

		// TODO: Implement multipart download for larger files
		go func(objectName string) error {
			defer wg.Done()

			reader, err := bucket.Object(objectName).NewReader(ctx)
			if err != nil {
				return fmt.Errorf("Error reading object [%s], err: [%v]", objectName, err)
			}
			defer reader.Close()

			// Create a local file to save the downloaded content
			filePath := filepath.Join(localFolder, objectName)
			outFile, err := os.Create(filePath)
			if err != nil {
				return fmt.Errorf("Error creating file [%s], Err:[%v]", filePath, err)
			}
			defer outFile.Close()

			// Copy the content from the GCS object to the local file
			if _, err := io.Copy(outFile, reader); err != nil {
				return fmt.Errorf("io.Copy: %v", err)
			}

			return nil
		}(objAttrs.Name)
	}

	return nil
}

func GcrUpload(ctx context.Context,
	bucketName, folderName string,
	wg *sync.WaitGroup, sem chan struct{}) error {

	defer wg.Done()

	// Initialize GCS client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("Failed to create GCS client: %v", err)
	}
	defer client.Close()

	// Walk through the folder and upload files concurrently
	err = filepath.Walk(folderName, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		sem <- struct{}{}
		wg.Add(1)
		go func(filePath string) error {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore slot

			relPath, err := filepath.Rel(folderName, filePath)
			if err != nil {
				return fmt.Errorf("Failed to get relative path: %v", err)
			}

			if err := uploadFileToGCS(ctx, bucketName, filePath, relPath); err != nil {
				return fmt.Errorf("Failed to upload %s: %v", filePath, err)
			} else {
				fmt.Printf("Uploaded %s to gs://%s/%s\n", filePath, bucketName, relPath)
			}
			return nil
		}(filePath)

		return nil
	})

	return err
}

func uploadFileToGCS(ctx context.Context, bucketName, filePath, gcsObjectName string) error {

	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %v", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	obj := bucket.Object(gcsObjectName)
	wc := obj.NewWriter(ctx)

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Upload file
	_, err = wc.Write([]byte{})
	if err != nil {
		return fmt.Errorf("failed to write to GCS: %w", err)
	}

	// Close writer
	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	fmt.Printf("Uploaded %s to gs://%s/%s\n", filePath, bucketName, gcsObjectName)
	return nil
}
