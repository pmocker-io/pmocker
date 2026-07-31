package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pmocker",
	Short: "PMocker - Docker for Project Management Systems",
	Long:  "PMocker 将项目管理系统封装为可分享的 .pmi 镜像，一条命令启动完整的项目管理系统。",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
