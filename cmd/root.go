package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/fatih/color"
	"moul.io/banner"
)

var rootCmd = &cobra.Command{
	Use:   "storage-synk",
	Short: "sorage-synk is a simple tool to transfer data to and fro between local and S3/GCP buckets",
	Run: func(cmd *cobra.Command, args []string) {
		color.Green("Welcome To:")
		color.Cyan(banner.Inline("storage-synk"))	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
