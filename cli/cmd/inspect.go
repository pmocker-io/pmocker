package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pmocker-io/pmocker/cli/internal/instance"
	"github.com/pmocker-io/pmocker/pkg/pmocker/image"
	"github.com/pmocker-io/pmocker/pkg/pmocker/oci"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <image|name:tag|instance>",
	Short: "查看镜像或实例的详细信息",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := args[0]

		// 1. 先尝试作为实例查询
		store, err := instance.InitDefaultStore()
		if err == nil {
			inst, err := store.GetByID(target)
			if err != nil {
				inst, err = store.GetByName(target)
			}
			if err == nil {
				store.Close()
				return printInstanceInfo(inst)
			}
			store.Close()
		}

		// 2. 尝试作为 .pmi 文件打开
		if _, err := os.Stat(target); err == nil {
			return inspectPMIFile(target)
		}

		// 3. 尝试从镜像缓存解析
		dir, _ := image.DefaultStoreDir()
		store2 := image.NewStore(dir)
		name, tag := parseImageRef(target)
		info, err := store2.ResolveImage(name, tag)
		if err != nil {
			return fmt.Errorf("无法解析 %s: %w", target, err)
		}
		return printImageInfo(info)
	},
}

func printInstanceInfo(inst *instance.Instance) error {
	fmt.Println("实例信息:")
	fmt.Printf("  ID:          %s\n", inst.ID)
	fmt.Printf("  名称:        %s\n", inst.Name)
	fmt.Printf("  镜像:        %s\n", inst.ImageRef)
	fmt.Printf("  镜像Digest: %s\n", inst.ImageDigest)
	fmt.Printf("  端口:        %d\n", inst.Port)
	fmt.Printf("  数据卷ID:   %s\n", inst.VolumeID)
	fmt.Printf("  状态:        %s\n", inst.Status)
	fmt.Printf("  PID:         %d\n", inst.PID)
	fmt.Printf("  创建时间:    %s\n", inst.CreatedAt.Format("2006-01-02 15:04:05"))
	if inst.StartedAt != nil {
		fmt.Printf("  启动时间:    %s\n", inst.StartedAt.Format("2006-01-02 15:04:05"))
	}
	if inst.StoppedAt != nil {
		fmt.Printf("  停止时间:    %s\n", inst.StoppedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
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
	cfgJSON, _ := json.MarshalIndent(info.Config, "", "  ")
	fmt.Println("Config:")
	fmt.Println(string(cfgJSON))
	manJSON, _ := json.MarshalIndent(info.Manifest, "", "  ")
	fmt.Println("\nManifest:")
	fmt.Println(string(manJSON))
	return nil
}

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
