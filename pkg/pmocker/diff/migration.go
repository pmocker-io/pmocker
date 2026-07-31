package diff

import "fmt"

// MigrationPlan 迁移计划
type MigrationPlan struct {
	Operations []MigrationOp `json:"operations"`
	Summary    string        `json:"summary"`
}

// MigrationOp 单个迁移操作
type MigrationOp struct {
	Type        string `json:"type"` // add_entity_type / remove_entity_type / add_field / remove_field / modify_field
	EntityType  string `json:"entity_type"`
	FieldKey    string `json:"field_key,omitempty"`
	DataType    string `json:"data_type,omitempty"`
	OldValue    string `json:"old_value,omitempty"`
	NewValue    string `json:"new_value,omitempty"`
	Description string `json:"description"`
}

// GenerateMigration 从 diff 结果生成迁移计划
func GenerateMigration(d *DiffResult) *MigrationPlan {
	plan := &MigrationPlan{}
	// 新增实体类型
	for _, et := range d.Schema.AddedEntityTypes {
		plan.Operations = append(plan.Operations, MigrationOp{
			Type: "add_entity_type", EntityType: et,
			Description: fmt.Sprintf("新增实体类型: %s", et),
		})
	}
	// 删除实体类型
	for _, et := range d.Schema.RemovedEntityTypes {
		plan.Operations = append(plan.Operations, MigrationOp{
			Type: "remove_entity_type", EntityType: et,
			Description: fmt.Sprintf("删除实体类型: %s", et),
		})
	}
	// 字段变更
	for _, fc := range d.Schema.FieldChanges {
		op := MigrationOp{
			EntityType: fc.EntityType,
			FieldKey:   fc.FieldKey,
		}
		switch fc.ChangeType {
		case "added":
			op.Type = "add_field"
			op.Description = fmt.Sprintf("新增字段: %s.%s", fc.EntityType, fc.FieldKey)
		case "removed":
			op.Type = "remove_field"
			op.Description = fmt.Sprintf("删除字段: %s.%s", fc.EntityType, fc.FieldKey)
		case "modified":
			op.Type = "modify_field"
			op.OldValue = fc.OldValue
			op.NewValue = fc.NewValue
			op.Description = fmt.Sprintf("修改字段: %s.%s (%s: %s -> %s)", fc.EntityType, fc.FieldKey, fc.ChangedAttr, fc.OldValue, fc.NewValue)
		}
		plan.Operations = append(plan.Operations, op)
	}
	total := len(plan.Operations)
	plan.Summary = fmt.Sprintf("共 %d 个迁移操作（新增 %d 实体, 删除 %d 实体, %d 字段变更）",
		total, len(d.Schema.AddedEntityTypes), len(d.Schema.RemovedEntityTypes), len(d.Schema.FieldChanges))
	return plan
}
