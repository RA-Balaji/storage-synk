package gcp

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
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
