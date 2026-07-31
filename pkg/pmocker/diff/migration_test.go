package diff

import (
	"testing"
)

func TestGenerateMigration(t *testing.T) {
	d := &DiffResult{
		Schema: SchemaDiff{
			AddedEntityTypes:   []string{"issue"},
			RemovedEntityTypes: []string{"risk"},
			FieldChanges: []FieldChange{
				{EntityType: "requirement", FieldKey: "priority", ChangeType: "added"},
				{EntityType: "requirement", FieldKey: "title", ChangeType: "modified", ChangedAttr: "data_type", OldValue: "string", NewValue: "text"},
			},
		},
	}
	plan := GenerateMigration(d)
	if len(plan.Operations) != 4 {
		t.Fatalf("operations = %d, want 4", len(plan.Operations))
	}
	// 检查操作类型
	types := map[string]int{}
	for _, op := range plan.Operations {
		types[op.Type]++
	}
	if types["add_entity_type"] != 1 {
		t.Errorf("add_entity_type = %d", types["add_entity_type"])
	}
	if types["remove_entity_type"] != 1 {
		t.Errorf("remove_entity_type = %d", types["remove_entity_type"])
	}
	if types["add_field"] != 1 {
		t.Errorf("add_field = %d", types["add_field"])
	}
	if types["modify_field"] != 1 {
		t.Errorf("modify_field = %d", types["modify_field"])
	}
}
