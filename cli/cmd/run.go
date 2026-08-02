package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pmocker-io/pmocker/cli/internal/builder"
	"github.com/pmocker-io/pmocker/cli/internal/instance"
	"github.com/pmocker-io/pmocker/pkg/pmocker/image"
	"github.com/spf13/cobra"
)

// findProjectRoot 基于 CLI 可执行文件位置推导项目根目录。
// 结构约定：<root>/cli/pmocker.exe → 上一级即为 root，内含 gva/server 和 gva/web。
func findProjectRoot() string {
	exe, err := os.Executable()
	if err == nil {
		// <root>/cli/pmocker.exe → <root>/cli → <root>
		root := filepath.Dir(filepath.Dir(exe))
		if _, err2 := os.Stat(filepath.Join(root, "gva", "server")); err2 == nil {
			if _, err3 := os.Stat(filepath.Join(root, "gva", "web")); err3 == nil {
				return root
			}
		}
	}
	wd, _ := os.Getwd()
	return wd
}

var runFollow bool
var runRebuild bool
var runForce bool

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "启动一个 PMSystem 实例",
	RunE: func(cmd *cobra.Command, args []string) error {
		imageRef, _ := cmd.Flags().GetString("image")
		portStr, _ := cmd.Flags().GetString("port")
		name, _ := cmd.Flags().GetString("name")
		adminPwd, _ := cmd.Flags().GetString("admin-password")

		if imageRef == "" {
			return fmt.Errorf("--image (-i) is required")
		}
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("invalid port: %s", portStr)
		}

		// 1. 确保二进制存在（--rebuild 时强制重建前后端，确保使用最新代码）
		pmockerHome, _ := instance.DefaultPMockerHome()
	gvaServerDir := findGVAServerDir()
		gvaWebDir := findGVAWebDir()
		b := builder.NewBuilder(pmockerHome, gvaServerDir, gvaWebDir)
		if runRebuild {
			if err := b.Rebuild(); err != nil {
				return fmt.Errorf("rebuild binary: %w", err)
			}
		} else {
			if err := b.Ensure(); err != nil {
				return fmt.Errorf("ensure binary: %w", err)
			}
		}

		// 2. 初始化存储
		store, err := instance.InitDefaultStore()
		if err != nil {
			return err
		}
		defer store.Close()

		vols, err := instance.InitDefaultVolumes()
		if err != nil {
			return err
		}

		imgDir, err := image.DefaultStoreDir()
		if err != nil {
			return err
		}
		imgStore := image.NewStore(imgDir)

		// 3. 创建实例管理器并运行
		mgr := instance.NewManager(store, vols, imgStore, b.BinPath())
		inst, err := mgr.Run(instance.RunOptions{
			ImageRef:      imageRef,
			Name:          name,
			Port:          port,
			AdminPassword: adminPwd,
			Force:         runForce,
		})
		if err != nil {
			return err
		}

		// 4. 输出结果
		fmt.Printf("实例已启动:\n")
		fmt.Printf("  ID:     %s\n", inst.ID)
		fmt.Printf("  名称:   %s\n", inst.Name)
		fmt.Printf("  镜像:   %s\n", inst.ImageRef)
		fmt.Printf("  端口:   %d\n", inst.Port)
		fmt.Printf("  状态:   %s\n", inst.Status)
		fmt.Printf("  PID:    %d\n", inst.PID)
		fmt.Printf("\n访问地址: http://localhost:%d\n", inst.Port)

		// 5. --follow 模式：自动跟踪 gva-server 实时日志，Ctrl+C 退出
		if runFollow {
			logPath := filepath.Join(vols.Path(inst.VolumeID), "gva-server.log")
			fmt.Printf("\n实时日志（Ctrl+C 退出，实例继续运行）:\n")
			fmt.Printf("----------------------------------------\n")
			// 等待日志文件创建（进程刚启动可能尚未写日志）
			for i := 0; i < 50; i++ {
				if _, err := os.Stat(logPath); err == nil {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
			return followLog(logPath, false)
		}
		return nil
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
	runCmd.Flags().BoolVarP(&runFollow, "follow", "f", false, "启动后自动跟踪 gva-server 实时日志（Ctrl+C 退出，实例继续运行）")
	runCmd.Flags().BoolVar(&runRebuild, "rebuild", false, "强制重建后端二进制和前端 dist，确保使用最新代码")
	runCmd.Flags().BoolVar(&runForce, "force", false, "同名实例存在时自动删除后重建")
}

func findGVAServerDir() string {
	return filepath.Join(findProjectRoot(), "gva", "server")
}

func findGVAWebDir() string {
	return filepath.Join(findProjectRoot(), "gva", "web")
}
