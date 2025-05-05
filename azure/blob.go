package azure

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func newServiceClient(ctx context.Context, connStr string) (*azblob.Client, error) {
	accountName, ok := os.LookupEnv("AZURE_STORAGE_ACCOUNT_NAME")
	if !ok {
		panic("AZURE_STORAGE_ACCOUNT_NAME could not be found")
	}

	accountKey, ok := os.LookupEnv("AZURE_STORAGE_PRIMARY_ACCOUNT_KEY")
	if !ok {
		panic("AZURE_STORAGE_PRIMARY_ACCOUNT_KEY could not be found")
	}
	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	handleError(err)

	// The service URL for blob endpoints is usually in the form: http(s)://<account>.blob.core.windows.net/
	client, err := azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net/", accountName), cred, nil)
	handleError(err)
	return client, nil

}

// BlobContainerCreate creates a new container.
func BlobContainerCreate(ctx context.Context, connStr, containerName string) error {
	svc, err := newServiceClient(ctx, connStr)
	if err != nil {
		return fmt.Errorf("initializing service client: %w", err)
	}
	container := svc.NewContainerClient(containerName)
	_, err = container.Create(ctx, nil)
	if err != nil {
		return fmt.Errorf("creating container %q: %w", containerName, err)
	}
	return nil
}

// BlobContainerGet returns true if the container exists (and its properties).
func BlobContainerGet(ctx context.Context, connStr, containerName string) (*azblob.ContainerGetPropertiesResponse, error) {
	svc, err := newServiceClient(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("initializing service client: %w", err)
	}
	container := svc.NewContainerClient(containerName)
	props, err := container.GetProperties(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("getting properties for container %q: %w", containerName, err)
	}
	return props, nil
}

// BlobContainerDelete deletes a container.
func BlobContainerDelete(ctx context.Context, connStr, containerName string) error {
	svc, err := newServiceClient(ctx, connStr)
	if err != nil {
		return fmt.Errorf("initializing service client: %w", err)
	}
	container := svc.NewContainerClient(containerName)
	_, err = container.Delete(ctx, nil)
	if err != nil {
		return fmt.Errorf("deleting container %q: %w", containerName, err)
	}
	return nil
}

// BlobFileDownload downloads a single blob to a local file path.
func BlobFileDownload(ctx context.Context, connStr, containerName, blobName, destPath string) error {
	svc, err := newServiceClient(ctx, connStr)
	if err != nil {
		return fmt.Errorf("initializing service client: %w", err)
	}
	container := svc.NewContainerClient(containerName)
	blob := container.NewBlobClient(blobName)

	resp, err := blob.DownloadStream(ctx, nil)
	if err != nil {
		return fmt.Errorf("downloading blob %q: %w", blobName, err)
	}
	defer resp.Body.Close()

	// ensure local dir exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("mkdir %q: %w", filepath.Dir(destPath), err)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("creating file %q: %w", destPath, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("writing to %q: %w", destPath, err)
	}
	return nil
}

// BlobFolderDownload walks all blobs under a given prefix (or root if prefix="")
// and downloads them in parallel, constrained by sem (e.g. make(chan struct{}, 10)).
// wg should be an existing *sync.WaitGroup; when the walk returns, call wg.Wait().
func BlobFolderDownload(
	ctx context.Context,
	connStr, containerName, prefix, outDir string,
	wg *sync.WaitGroup, sem chan struct{},
) error {
	svc, err := newServiceClient(ctx, connStr)
	if err != nil {
		return fmt.Errorf("initializing service client: %w", err)
	}
	container := svc.NewContainerClient(containerName)
	pager := container.NewListBlobsFlatPager(&azblob.ListBlobsFlatOptions{
		Prefix: &prefix,
	})

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("listing blobs: %w", err)
		}
		for _, blobItem := range page.Segment.BlobItems {
			name := *blobItem.Name
			wg.Add(1)
			sem <- struct{}{} // acquire

			go func(blobName string) {
				defer wg.Done()
				defer func() { <-sem }() // release

				// build a local file path from the blob name, preserving folders
				localPath := filepath.Join(outDir, filepath.FromSlash(blobName))
				if err := BlobFileDownload(ctx, connStr, containerName, blobName, localPath); err != nil {
					fmt.Fprintf(os.Stderr, "error downloading %q: %v\n", blobName, err)
				}
			}(name)
		}
	}

	return nil
}

// isWindows returns true if running on Windows.
func isWindows() bool {
	return runtime.GOOS == "windows"
}

// convToLocalPath converts forward‐slash blob names into OS‐native separators.
// (Note: filepath.FromSlash does this for you, so you usually won’t need it.)
func convToLocalPath(blobName string) string {
	if isWindows() {
		return strings.ReplaceAll(blobName, "/", "\\")
	}
	return blobName
}
