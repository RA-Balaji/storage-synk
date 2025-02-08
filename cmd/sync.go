package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sync"

	//awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/RA-Balaji/storage-synk/aws"
	"github.com/RA-Balaji/storage-synk/gcp"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/spf13/cobra"
)

var csp, source, dest string

const (
	cspGcp = "gcp"
	cspAws = "aws"
)

var cpCmd = &cobra.Command{
	Use:   "cp",
	Short: "copies files/folder between source and destination",
	RunE: func(cmd *cobra.Command, args []string) error {
		source, err := cmd.Flags().GetString("source")
		if err != nil {
			return fmt.Errorf("Source incorrect, error: %v", err)
		}
		destination, err := cmd.Flags().GetString("destination")
		if err != nil {
			return fmt.Errorf("Destination incorrect, error: %v", err)
		}
		tmpPath, err := cmd.Flags().GetString("destination")
		if err != nil {
			return fmt.Errorf("Error loading temp path: %v", err)
		}
		awsProfile, err := cmd.Flags().GetString("aws-profile")
		if err != nil {
			return fmt.Errorf("Error parsing aws-profile: %v", err)
		}

		err = validateSrcDst(source, destination)
		if err != nil {
			return err
		}

		ctx := context.Background()
		if source == cspGcp && destination == cspAws {
			err = TransferFromGcpToAWS(
				ctx, awsProfile, source, destination, tmpPath)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cpCmd)

	cpCmd.Flags().StringP("source", "s", "", "Source bucket path")
	cpCmd.Flags().StringP("destination", "d", "", "Destination bucket path")
	cpCmd.PersistentFlags().Lookup("download-location").NoOptDefVal = os.TempDir()
	cpCmd.PersistentFlags().Lookup("aws-profile").NoOptDefVal = "default"
}

func validateSrcDst(src, dst string) error {
	var validGCPBucketPath = regexp.MustCompile(`^gs://[a-zA-Z0-9._-]+(/[a-zA-Z0-9._-]+)*$`)
	var validS3Path = regexp.MustCompile(`^s3://[a-zA-Z0-9._-]+(/[a-zA-Z0-9._-]+)*$`)

	if !validGCPBucketPath.MatchString(src) && !validS3Path.MatchString(src) {
		return fmt.Errorf("Invalid Source: %s", source)
	}

	if !validGCPBucketPath.MatchString(dst) && !validS3Path.MatchString(dst) {
		return fmt.Errorf("Invalid Destination: %s", dst)
	}

	return nil
}

func TransferFromGcpToAWS(
	ctx context.Context,
	profile,
	source, destination, tmpPath string) error {
	err := gcp.GcsDownload(ctx, source, tmpPath)
	if err != nil {
		return err
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithSharedConfigProfile(profile))
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // TODO: allow user to configure the concurrency limit?

	err = aws.S3FolderUpload(ctx, cfg.Region, destination, tmpPath, &wg, sem)
	if err != nil {
		return err
	}

	// Wait for all uploads to finish
	wg.Wait()
	fmt.Println("Folder upload completed successfully!")

	return nil
}
