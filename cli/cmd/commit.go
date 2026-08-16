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
	Short: "从 PMSystem 实例导出新镜像（含实例当前数据）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target, _ := cmd.Flags().GetString("tag")
		message, _ := cmd.Flags().GetString("message")
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

		vols, err := instance.InitDefaultVolumes()
		if err != nil {
			return err
		}
		volPath := vols.Path(inst.VolumeID)
		if _, err := os.Stat(volPath); os.IsNotExist(err) {
			return fmt.Errorf("instance volume not found: %s", volPath)
		}

		// 定位原镜像
		imgDir, _ := image.DefaultStoreDir()
		origPath := filepath.Join(imgDir, strings.ReplaceAll(inst.ImageDigest, ":", "_"), "image.pmi")
		if _, err := os.Stat(origPath); os.IsNotExist(err) {
			return fmt.Errorf("original image not found: %s", origPath)
		}

		// 构建新镜像（原层 + 实例 data 层）
		outPath := filepath.Join(os.TempDir(), "pmocker-commit-"+inst.ID+".pmi")
		defer os.Remove(outPath)
		if err := instance.BuildCommitImage(origPath, volPath, outPath); err != nil {
			return fmt.Errorf("build commit image: %w", err)
		}

		// 注册到镜像库
		name, tag := parseImageRef(target)
		imgStore := image.NewStore(imgDir)
		info, err := imgStore.AddImage(outPath, name, tag)
		if err != nil {
			return fmt.Errorf("add image: %w", err)
		}

		fmt.Printf("已提交新镜像:\n")
		fmt.Printf("  名称:   %s\n", name)
		fmt.Printf("  标签:   %s\n", tag)
		fmt.Printf("  Digest: %s\n", info.Digest)
		if message != "" {
			fmt.Printf("  说明:   %s\n", message)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
	commitCmd.Flags().StringP("tag", "t", "", "新镜像名:tag")
	commitCmd.Flags().StringP("message", "m", "", "提交说明")
}
