package diff

import (
	"path/filepath"
	"testing"

	"github.com/pmocker-io/pmocker/pkg/pmocker/oci"
	"github.com/pmocker-io/pmocker/pkg/pmocker/plugin/loader"
)

func TestDiffSchemas(t *testing.T) {
	old := []loader.EntitySchemaYaml{
		{EntityType: "requirement", Fields: []loader.FieldYaml{
			{FieldKey: "title", DataType: "string"},
			{FieldKey: "priority", DataType: "enum", OptionsJSON: "[\"high\",\"low\"]"},
		}},
		{EntityType: "risk", Fields: []loader.FieldYaml{
			{FieldKey: "level", DataType: "string"},
		}},
	}
	new := []loader.EntitySchemaYaml{
		{EntityType: "requirement", Fields: []loader.FieldYaml{
			{FieldKey: "title", DataType: "string"},
			{FieldKey: "priority", DataType: "enum", OptionsJSON: "[\"high\",\"medium\",\"low\"]"},
			{FieldKey: "assignee", DataType: "ref"},
		}},
		{EntityType: "issue", Fields: []loader.FieldYaml{
			{FieldKey: "severity", DataType: "enum"},
		}},
	}
	d := DiffSchemas(old, new)
	if len(d.AddedEntityTypes) != 1 || d.AddedEntityTypes[0] != "issue" {
		t.Errorf("added entity types = %v", d.AddedEntityTypes)
	}
	if len(d.RemovedEntityTypes) != 1 || d.RemovedEntityTypes[0] != "risk" {
		t.Errorf("removed entity types = %v", d.RemovedEntityTypes)
	}
	addedCount := 0
	modifiedCount := 0
	for _, c := range d.FieldChanges {
		if c.ChangeType == "added" {
			addedCount++
		}
		if c.ChangeType == "modified" {
			modifiedCount++
		}
	}
	if addedCount != 1 {
		t.Errorf("added fields = %d, want 1", addedCount)
	}
	if modifiedCount != 1 {
		t.Errorf("modified fields = %d, want 1", modifiedCount)
	}
}

func TestDiffImages(t *testing.T) {
	tmpDir := t.TempDir()
	oldPath := filepath.Join(tmpDir, "old.pmi")
	newPath := filepath.Join(tmpDir, "new.pmi")

	// 构建旧镜像
	oldFiles := map[string][]byte{
		"requirement.yaml": []byte("entity_type: requirement\nfields:\n  - {field_key: title, data_type: string}\n"),
	}
	_, oldTar, _ := oci.CreateLayerFromFiles(oldFiles, oci.LayerTypeSchema)
	oci.BuildImage(oldPath, oci.NewConfig("Old", "1.0.0", []string{"requirement"}), []oci.LayerData{oci.NewLayerData(oldTar, oci.LayerTypeSchema)})

	// 构建新镜像（多一个字段）
	newFiles := map[string][]byte{
		"requirement.yaml": []byte("entity_type: requirement\nfields:\n  - {field_key: title, data_type: string}\n  - {field_key: priority, data_type: enum}\n"),
	}
	_, newTar, _ := oci.CreateLayerFromFiles(newFiles, oci.LayerTypeSchema)
	oci.BuildImage(newPath, oci.NewConfig("New", "2.0.0", []string{"requirement"}), []oci.LayerData{oci.NewLayerData(newTar, oci.LayerTypeSchema)})

	result, err := DiffImages(oldPath, newPath)
	if err != nil {
		t.Fatalf("DiffImages: %v", err)
	}
	if len(result.Schema.FieldChanges) != 1 {
		t.Errorf("field changes = %d, want 1", len(result.Schema.FieldChanges))
	}
	if result.Schema.FieldChanges[0].ChangeType != "added" {
		t.Errorf("change type = %s", result.Schema.FieldChanges[0].ChangeType)
	}
}
