package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newStubCmd 创建占位命令（M5 阶段实现具体逻辑）
func newStubCmd(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("TODO: pmocker %s — M5 实现\n", use)
		},
	}
}

func init() {
	// 实例管理
	rootCmd.AddCommand(newStubCmd("ps", "列出运行中的 PMSystem"))
	rootCmd.AddCommand(newStubCmd("start", "启动已停止的 PMSystem"))
	rootCmd.AddCommand(newStubCmd("stop", "停止 PMSystem"))
	rootCmd.AddCommand(newStubCmd("rm", "删除 PMSystem 实例"))

	// 镜像管理
	rootCmd.AddCommand(newStubCmd("commit", "从 PMSystem 导出新镜像"))
	rootCmd.AddCommand(newStubCmd("export", "导出 PMSystem 为 .pmi 文件"))
	rootCmd.AddCommand(newStubCmd("images", "列出本地镜像"))
	rootCmd.AddCommand(newStubCmd("inspect", "查看镜像或实例详情"))
	rootCmd.AddCommand(newStubCmd("rmi", "删除本地镜像"))

	// 升级
	rootCmd.AddCommand(newStubCmd("diff", "对比两个 .pmi 镜像差异"))
	rootCmd.AddCommand(newStubCmd("upgrade", "升级 PMSystem 到新镜像版本"))
}
