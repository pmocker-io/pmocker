package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "启动一个 PMSystem 实例",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("TODO: pmocker run — M5 实现")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringP("image", "i", "", "指定 .pmi 镜像文件或镜像名")
	runCmd.Flags().StringP("port", "p", "8080", "暴露端口")
	runCmd.Flags().StringP("name", "n", "", "实例名称")
	runCmd.Flags().StringP("db", "d", "sqlite", "数据库驱动 (sqlite|mysql|postgres)")
	runCmd.Flags().String("db-dsn", "", "数据库 DSN（mysql/postgres 时必填）")
	runCmd.Flags().StringP("volume", "v", "", "数据卷路径")
	runCmd.Flags().String("admin-password", "", "管理员密码")
}
