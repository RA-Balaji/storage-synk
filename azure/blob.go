package azure

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func newServiceClient(ctx context.Context, connectionString string) *azblob.Client {
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	handleError(err)
	return client
}

// BlobContainerCreate creates a new container.
func BlobContainerCreate(ctx context.Context, connStr, containerName string) {
	client := newServiceClient(ctx, connStr)
	_, err := client.CreateContainer(context.TODO(), containerName, nil)
	handleError(err)
}

// BlobContainerGet returns true if the container exists (and its properties).
func BlobContainerGet(ctx context.Context, connStr, containerName string) container.GetPropertiesResponse {
	client := newServiceClient(ctx, connStr)
	containerClient := client.ServiceClient().NewContainerClient(containerName)

	// Get the container properties
	props, err := containerClient.GetProperties(context.TODO(), nil)
	handleError(err)
	return props
}

// BlobContainerDelete deletes a container.
func BlobContainerDelete(ctx context.Context, connStr, containerName string) {
	client := newServiceClient(ctx, connStr)
	_, err := client.DeleteContainer(context.TODO(), containerName, nil)
	handleError(err)
}

// BlobFileDownload downloads a single blob to a local file path.
func BlobFileDownload(ctx context.Context, connStr, containerName, blobName, destPath string) {
	client := newServiceClient(ctx, connStr)
	file, err := os.Create(destPath + "/" + blobName)
	handleError(err)
	_, err = client.DownloadFile(context.TODO(), containerName, blobName, file, nil)
	handleError(err)
}

// BlobFolderDownload walks all blobs under a given prefix (or root if prefix="")
// and downloads them in parallel, constrained by sem (e.g. make(chan struct{}, 10)).
// wg should be an existing *sync.WaitGroup; when the walk returns, call wg.Wait().
func BlobFolderDownload(
	ctx context.Context,
	connStr, containerName, prefix, destBasePath string,
) error {
	// 1) Make sure dest base exists
	if err := os.MkdirAll(destBasePath, 0o755); err != nil {
		return fmt.Errorf("mkdir %q: %w", destBasePath, err)
	}

	// 2) List blobs
	client := newServiceClient(ctx, connStr)

	pager := client.NewListBlobsFlatPager(containerName, &azblob.ListBlobsFlatOptions{
		Prefix: &prefix,
	})

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("listing blobs: %w", err)
		}

		for _, b := range page.Segment.BlobItems {
			blobName := *b.Name
			// build the local file path
			localPath := filepath.Join(destBasePath, filepath.FromSlash(blobName))

			// ensure parent dirs exist
			if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
				return fmt.Errorf("mkdir parent for %q: %w", localPath, err)
			}

			// and download
			BlobFileDownload(ctx, connStr, containerName, blobName, destBasePath)
		}
	}

	return nil
}

// // isWindows returns true if running on Windows.
// func isWindows() bool {
// 	return runtime.GOOS == "windows"
// }

// // convToLocalPath converts forward‐slash blob names into OS‐native separators.
// // (Note: filepath.FromSlash does this for you, so you usually won’t need it.)
// func convToLocalPath(blobName string) string {
// 	if isWindows() {
// 		return strings.ReplaceAll(blobName, "/", "\\")
// 	}
// 	return blobName
// }
