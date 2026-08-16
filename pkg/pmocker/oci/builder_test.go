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

// TestBuildImageWithDataLayer 构建含 data 层（实例快照）的镜像并回读验证
func TestBuildImageWithDataLayer(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "withdata.pmi")

	cfg := NewConfig("TestMethod", "2.0.0", []string{"requirement"})

	schemaFiles := map[string][]byte{
		"requirement.yaml": []byte("entity_type: requirement\nfields: []"),
	}
	_, schemaTar, _ := CreateLayerFromFiles(schemaFiles, LayerTypeSchema)
	schemaLayer := NewLayerData(schemaTar, LayerTypeSchema)

	// data 层：模拟实例数据卷（system.db + dist）
	dataFiles := map[string][]byte{
		"system.db":  []byte("fake sqlite bytes"),
		"dist/index.html": []byte("<html>pms</html>"),
	}
	_, dataTar, _ := CreateLayerFromFiles(dataFiles, LayerTypeData)
	dataLayer := NewLayerData(dataTar, LayerTypeData)

	if err := BuildImage(outPath, cfg, []LayerData{schemaLayer, dataLayer}); err != nil {
		t.Fatalf("BuildImage: %v", err)
	}

	// 回读验证
	r, err := OpenImage(outPath)
	if err != nil {
		t.Fatalf("OpenImage: %v", err)
	}

	m := r.Manifest()
	if len(m.Layers) != 2 {
		t.Fatalf("layers = %d, want 2", len(m.Layers))
	}
	// data 层应可提取
	data, err := r.ExtractLayerByType(LayerTypeData)
	if err != nil {
		t.Fatalf("ExtractLayerByType(data): %v", err)
	}
	files, err := r.ExtractLayerFiles(LayerTypeData)
	if err != nil {
		t.Fatalf("ExtractLayerFiles(data): %v", err)
	}
	if _, ok := files["system.db"]; !ok {
		t.Error("data 层缺少 system.db")
	}
	_ = data
}
