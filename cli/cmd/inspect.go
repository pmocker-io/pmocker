package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pmocker-io/pmocker/pkg/pmocker/image"
	"github.com/pmocker-io/pmocker/pkg/pmocker/oci"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <image|name:tag>",
	Short: "查看镜像或实例的详细信息",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]
		// 尝试作为 .pmi 文件打开
		if _, err := os.Stat(target); err == nil {
			return inspectPMIFile(target)
		}
		// 尝试从镜像缓存解析
		dir, _ := image.DefaultStoreDir()
		store := image.NewStore(dir)
		// 尝试 name:tag 或 name
		name, tag := parseImageRef(target)
		info, err := store.ResolveImage(name, tag)
		if err != nil {
			return fmt.Errorf("无法解析 %s: %w", target, err)
		}
		return printImageInfo(info)
	},
}

func inspectPMIFile(path string) error {
	reader, err := oci.OpenImage(path)
	if err != nil {
		return err
	}
	info := &image.ImageInfo{
		Digest:   reader.Manifest().Config.Digest,
		Config:   reader.Config(),
		Manifest: reader.Manifest(),
	}
	return printImageInfo(info)
}

func printImageInfo(info *image.ImageInfo) error {
	// 打印 config
	cfgJSON, _ := json.MarshalIndent(info.Config, "", "  ")
	fmt.Println("Config:")
	fmt.Println(string(cfgJSON))
	// 打印 manifest
	manJSON, _ := json.MarshalIndent(info.Manifest, "", "  ")
	fmt.Println("\nManifest:")
	fmt.Println(string(manJSON))
	return nil
}

// parseImageRef 解析镜像引用 name:tag，缺省 tag 为 latest
func parseImageRef(ref string) (name, tag string) {
	for i := len(ref) - 1; i >= 0; i-- {
		if ref[i] == ':' {
			return ref[:i], ref[i+1:]
		}
	}
	return ref, "latest"
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
