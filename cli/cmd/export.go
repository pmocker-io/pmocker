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
	Short: "导出 PMSystem 实例为 .pmi 文件",
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

		imgDir, _ := image.DefaultStoreDir()
		pmiPath := filepath.Join(imgDir, strings.ReplaceAll(inst.ImageDigest, ":", "_"), "image.pmi")
		if _, err := os.Stat(pmiPath); os.IsNotExist(err) {
			return fmt.Errorf("image file not found: %s", pmiPath)
		}

		data, err := os.ReadFile(pmiPath)
		if err != nil {
			return fmt.Errorf("read image: %w", err)
		}
		if err := os.WriteFile(output, data, 0644); err != nil {
			return fmt.Errorf("write output: %w", err)
		}

		fmt.Printf("已导出到: %s\n", output)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringP("output", "o", "", "输出文件路径")
}
