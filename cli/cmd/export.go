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

var exportCmd = &cobra.Command{
	Use:   "export <name|id> -o <file.pmi>",
	Short: "导出 PMSystem 实例为 .pmi 文件（含实例当前数据）",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			return fmt.Errorf("--output (-o) is required")
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

		imgDir, _ := image.DefaultStoreDir()
		origPath := filepath.Join(imgDir, strings.ReplaceAll(inst.ImageDigest, ":", "_"), "image.pmi")
		if _, err := os.Stat(origPath); os.IsNotExist(err) {
			return fmt.Errorf("original image not found: %s", origPath)
		}

		if err := instance.BuildCommitImage(origPath, volPath, output); err != nil {
			return fmt.Errorf("build export image: %w", err)
		}

		fmt.Printf("已导出到: %s\n", output)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringP("output", "o", "", "输出文件路径")
}
