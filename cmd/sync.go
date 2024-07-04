package cmd

import (
	"fmt"
	"regexp"

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

		_, err = validateSrcDst(source, destination)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cpCmd)

	cpCmd.Flags().StringP("source", "s", "", "Source bucket path")
	cpCmd.Flags().StringP("destination", "d", "", "Destination bucket path")
}

type SrcDst struct {
	Source      string
	Destination string
}

func validateSrcDst(src, dst string) (SrcDst, error) {
	var out SrcDst
	var validGCPBucketPath = regexp.MustCompile(`^gs://[a-zA-Z0-9._-]+(/[a-zA-Z0-9._-]+)*$`)
	var validS3Path = regexp.MustCompile(`^s3://[a-zA-Z0-9._-]+(/[a-zA-Z0-9._-]+)*$`)

	if validGCPBucketPath.MatchString(src) {
		out.Source = "gcp"
	} else if validS3Path.MatchString(src) {
		out.Source = "aws"
	} else {
		return out, fmt.Errorf("Invalid Source: %s", source)
	}

	return out, nil
}
