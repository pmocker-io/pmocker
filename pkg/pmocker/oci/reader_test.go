package oci

import (
	"path/filepath"
	"testing"
)

func TestOpenImage(t *testing.T) {
	// 先构建一个测试镜像
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "test.pmi")

	cfg := NewConfig("Test", "1.0.0", []string{"requirement"})
	schemaFiles := map[string][]byte{
		"requirement.yaml": []byte("entity_type: requirement\nfields: []"),
	}
	_, schemaTar, _ := CreateLayerFromFiles(schemaFiles, LayerTypeSchema)
	schemaLayer := NewLayerData(schemaTar, LayerTypeSchema)
	BuildImage(outPath, cfg, []LayerData{schemaLayer})

	// 读取
	r, err := OpenImage(outPath)
	if err != nil {
		t.Fatalf("OpenImage: %v", err)
	}
	if r.Config().Methodology != "Test" {
		t.Errorf("methodology = %s", r.Config().Methodology)
	}
	if r.Manifest().SchemaVersion != 2 {
		t.Errorf("schemaVersion = %d", r.Manifest().SchemaVersion)
	}

	// 提取 schema 层
	files, err := r.ExtractLayerFiles(LayerTypeSchema)
	if err != nil {
		t.Fatalf("ExtractLayerFiles: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("file count = %d, want 1", len(files))
	}
	if _, ok := files["requirement.yaml"]; !ok {
		t.Error("requirement.yaml not found in schema layer")
	}
}
