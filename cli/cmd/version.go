package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version 在构建时通过 -ldflags 注入
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示 PMocker 版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("PMocker %s\n", Version)
		fmt.Printf("  GitCommit: %s\n", GitCommit)
		fmt.Printf("  BuildDate: %s\n", BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
