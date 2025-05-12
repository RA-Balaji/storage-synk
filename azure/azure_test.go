package azure

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testContainerName = "surya-test-container"
	testBlobName      = "README.md"
	testPrefix        = ""
	connectionString  = ""
)

func TestContainerCreate(t *testing.T) {
	BlobContainerCreate(context.Background(), connectionString, testContainerName)
}

func TestContainerGet(t *testing.T) {
	props := BlobContainerGet(context.Background(), connectionString, testContainerName)
	log.Printf("Version: %s\n", *props.RequestID)
}

func TestContainerDelete(t *testing.T) {
	BlobContainerDelete(context.Background(), connectionString, testContainerName)
}

func TestBlobFileDownload(t *testing.T) {
	ctx := context.Background()
	dest := "/Users/suryakiransureshkumar/Downloads"

	BlobFileDownload(ctx, connectionString, testContainerName, testBlobName, dest)
	_, statErr := os.Stat(dest)
	assert.NoError(t, statErr)
	log.Printf("Downloaded file to: %s", dest)
}

func TestBlobFolderDownload(t *testing.T) {
	ctx := context.Background()
	outDir := filepath.Join(os.TempDir(), "azure-folder-download")

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

	err := BlobFolderDownload(ctx, "", testContainerName, testPrefix, outDir, &wg, sem)
	if err != nil {
		log.Printf("Folder download failed: %v", err)
		t.Fatal()
	}
	wg.Wait()

	filesDownloaded := 0
	_ = filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			filesDownloaded++
		}
		return nil
	})

	assert.Greater(t, filesDownloaded, 0)
	log.Printf("Downloaded %d files to %s", filesDownloaded, outDir)
}
