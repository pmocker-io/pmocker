package oci

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildImage(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "test.pmi")

	cfg := NewConfig("TestMethod", "1.0.0", []string{"requirement", "scope"})

	schemaFiles := map[string][]byte{
		"requirement.yaml": []byte("entity_type: requirement\nfields: []"),
		"scope.yaml":       []byte("entity_type: scope_item\nfields: []"),
	}
	_, schemaTar, _ := CreateLayerFromFiles(schemaFiles, LayerTypeSchema)
	schemaLayer := NewLayerData(schemaTar, LayerTypeSchema)

	pluginFiles := map[string][]byte{
		"pmocker_requirement/plugin.go": []byte("package pmocker_requirement"),
	}
	_, pluginTar, _ := CreateLayerFromFiles(pluginFiles, LayerTypePlugins)
	pluginLayer := NewLayerData(pluginTar, LayerTypePlugins)

	if err := BuildImage(outPath, cfg, []LayerData{schemaLayer, pluginLayer}); err != nil {
		t.Fatalf("BuildImage: %v", err)
	}

	// 验证文件存在
	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() == 0 {
		t.Error("image file is empty")
	}

	// 读取 tar 验证内容
	data, _ := os.ReadFile(outPath)
	tr := tar.NewReader(bytes.NewReader(data))
	files := map[string]bool{}
	for {
		hdr, err := tr.Next()
		if err != nil {
			break
		}
		files[hdr.Name] = true
		// 验证 manifest.json 可解析
		if hdr.Name == "manifest.json" {
			var m Manifest
			if err := json.NewDecoder(tr).Decode(&m); err != nil {
				t.Errorf("parse manifest: %v", err)
			}
			if m.SchemaVersion != 2 {
				t.Errorf("schemaVersion = %d", m.SchemaVersion)
			}
			if len(m.Layers) != 2 {
				t.Errorf("layers = %d, want 2", len(m.Layers))
			}
		}
	}
	if !files["manifest.json"] {
		t.Error("manifest.json not found")
	}
	if !files["config.json"] {
		t.Error("config.json not found")
	}
}
