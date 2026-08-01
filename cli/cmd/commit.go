package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pmocker-io/pmocker/cli/internal/instance"
	"github.com/pmocker-io/pmocker/pkg/pmocker/image"
	"github.com/spf13/cobra"
)

var commitCmd = &cobra.Command{
	Use:   "commit <name|id> -t <新镜像名:tag>",
	Short: "从 PMSystem 实例导出新镜像",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target, _ := cmd.Flags().GetString("tag")
		if target == "" {
			return fmt.Errorf("--tag (-t) is required")
		}

		store, err := instance.InitDefaultStore()
		if err != nil {
			return err
		}
		defer store.Close()

		inst, err := store.GetByID(args[0])
		if err != nil {
			inst, err = store.GetByName(args[0])
			if err != nil {
				return fmt.Errorf("instance %s not found", args[0])
			}
		}

		imgDir, _ := image.DefaultStoreDir()
		imgStore := image.NewStore(imgDir)

		name, tag := parseImageRef(target)
		origPath := filepath.Join(imgDir, strings.ReplaceAll(inst.ImageDigest, ":", "_"), "image.pmi")
		if _, err := os.Stat(origPath); os.IsNotExist(err) {
			return fmt.Errorf("original image file not found: %s", origPath)
		}

		info, err := imgStore.AddImage(origPath, name, tag)
		if err != nil {
			return fmt.Errorf("commit image: %w", err)
		}

		fmt.Printf("已提交新镜像:\n")
		fmt.Printf("  名称:   %s\n", name)
		fmt.Printf("  标签:   %s\n", tag)
		fmt.Printf("  Digest: %s\n", info.Digest)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
	commitCmd.Flags().StringP("tag", "t", "", "新镜像名:tag")
	commitCmd.Flags().StringP("message", "m", "", "提交说明")
}
