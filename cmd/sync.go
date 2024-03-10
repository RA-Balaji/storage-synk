package cmd

import (
	"github.com/spf13/cobra"
)

var csp, source, dest string

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "syncs files/folder between source and destination",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {

}
