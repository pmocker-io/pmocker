package cmd

import (
	"fmt"

	"github.com/pmocker-io/pmocker/pkg/pmocker/image"
	"github.com/spf13/cobra"
)

var rmiCmd = &cobra.Command{
	Use:   "rmi <image:tag|digest>",
	Short: "删除本地镜像缓存",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		dir, err := image.DefaultStoreDir()
		if err != nil {
			return err
		}
		store := image.NewStore(dir)
		// 如果是 digest 直接删除
		if len(target) > 7 && target[:7] == "sha256:" {
			if err := store.RemoveImage(target); err != nil {
				return err
			}
			fmt.Printf("已删除镜像: %s\n", target)
			return nil
		}
		// 否则按 name:tag 解析
		name, tag := parseImageRef(target)
		info, err := store.ResolveImage(name, tag)
		if err != nil {
			return err
		}
		if err := store.RemoveImage(info.Digest); err != nil {
			return err
		}
		fmt.Printf("已删除镜像: %s:%s (%s)\n", name, tag, info.Digest)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmiCmd)
}
