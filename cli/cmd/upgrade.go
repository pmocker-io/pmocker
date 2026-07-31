package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/pmocker-io/pmocker/pkg/pmocker/diff"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade <name|id> --to <新镜像>",
	Short: "将 PMSystem 升级到新镜像版本",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("需要指定实例名称或ID")
		}
		newImage, _ := cmd.Flags().GetString("to")
		if newImage == "" {
			return fmt.Errorf("需要指定 --to <新镜像>")
		}
		instance := args[0]
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// 获取实例的当前镜像路径（M4 简化：直接用文件路径）
		oldImage, _ := cmd.Flags().GetString("from")
		if oldImage == "" {
			return fmt.Errorf("需要指定 --from <旧镜像>（M5 将自动检测实例镜像）")
		}

		// 1. diff
		result, err := diff.DiffImages(oldImage, newImage)
		if err != nil {
			return fmt.Errorf("diff 失败: %w", err)
		}
		// 2. 生成 migration
		plan := diff.GenerateMigration(result)
		// 输出计划
		planJSON, _ := json.MarshalIndent(plan, "", "  ")
		fmt.Println("Migration Plan:")
		fmt.Println(string(planJSON))

		if dryRun {
			fmt.Println("\n--dry-run 模式，不执行迁移")
			return nil
		}
		// 3. 执行迁移（M4 打印操作，M5 连接数据库执行）
		fmt.Printf("\n正在升级实例 %s...\n", instance)
		for _, op := range plan.Operations {
			fmt.Printf("  [执行] %s: %s\n", op.Type, op.Description)
		}
		fmt.Printf("升级完成: %s → %s\n", oldImage, newImage)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().String("to", "", "目标新镜像路径")
	upgradeCmd.Flags().String("from", "", "当前镜像路径（M5 将自动检测）")
	upgradeCmd.Flags().Bool("dry-run", false, "仅打印迁移计划，不执行")
}
