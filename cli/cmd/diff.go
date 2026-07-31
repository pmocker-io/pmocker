package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/pmocker-io/pmocker/pkg/pmocker/diff"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff <旧镜像> <新镜像>",
	Short: "对比两个 .pmi 镜像的 schema/plugins/seed 差异",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		oldPath := args[0]
		newPath := args[1]
		result, err := diff.DiffImages(oldPath, newPath)
		if err != nil {
			return err
		}
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return nil
		}
		// 文本格式输出
		fmt.Println("=== Schema Diff ===")
		if len(result.Schema.AddedEntityTypes) > 0 {
			fmt.Printf("新增实体类型: %v\n", result.Schema.AddedEntityTypes)
		}
		if len(result.Schema.RemovedEntityTypes) > 0 {
			fmt.Printf("删除实体类型: %v\n", result.Schema.RemovedEntityTypes)
		}
		if len(result.Schema.FieldChanges) > 0 {
			fmt.Println("字段变更:")
			for _, c := range result.Schema.FieldChanges {
				switch c.ChangeType {
				case "added":
					fmt.Printf("  + %s.%s\n", c.EntityType, c.FieldKey)
				case "removed":
					fmt.Printf("  - %s.%s\n", c.EntityType, c.FieldKey)
				case "modified":
					fmt.Printf("  ~ %s.%s (%s: %s -> %s)\n", c.EntityType, c.FieldKey, c.ChangedAttr, c.OldValue, c.NewValue)
				}
			}
		}
		// 生成 migration 计划
		plan := diff.GenerateMigration(result)
		fmt.Printf("\n=== Migration Plan ===\n%s\n", plan.Summary)
		for _, op := range plan.Operations {
			fmt.Printf("  [%s] %s\n", op.Type, op.Description)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().BoolP("json", "j", false, "JSON 格式输出")
}
