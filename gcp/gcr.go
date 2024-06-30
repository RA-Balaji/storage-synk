package gcp

import (
	"context"
	"fmt"
	//"sync"

	"cloud.google.com/go/storage"
	//"google.golang.org/api/iterator"
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

// func GcsDownload(ctx context.Context, destination, bucketName string) error {
// 	client, err := storage.NewClient(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to create storage client: %v", err)
// 	}
// 	defer client.Close()

// 	bucket := client.Bucket(bucketName)

// 	var wg sync.WaitGroup
//     //var mu sync.Mutex

// 	it := bucket.Objects(ctx, nil)
//     for {
//         objAttrs, err := it.Next()
//         if err == iterator.Done {
//             break
//         }
//         if err != nil {
//             return fmt.Errorf("Error iterating Objects: %v", err)
//         }

//         // Increment the WaitGroup counter
//         wg.Add(1)

// 		go func(objectName string) {
// 			defer wg.Done()

// 			reader, err := bucket.Object(objectName).NewReader(ctx)
// 			if err != nil {

// 			}

// 		}(objAttrs.Name)
// 	}

// 	return nil
// }
