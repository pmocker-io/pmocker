package instance

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pmocker-io/pmocker/pkg/pmocker/oci"
)

// TestBuildCommitImage commit 镜像 = 原层 + data 层（含快照）
func TestBuildCommitImage(t *testing.T) {
	dir := t.TempDir()

	// 1. 构造原镜像（schema + plugin 层）
	origPath := filepath.Join(dir, "orig.pmi")
	cfg := oci.NewConfig("PMBOK6", "1.0.0", []string{"requirement"})
	schemaTar, _ := createTarFromFiles(map[string][]byte{
		"requirement.yaml": []byte("entity_type: requirement\n"),
	})
	pluginTar, _ := createTarFromFiles(map[string][]byte{
		"plugin.go": []byte("package x"),
	})
	if err := oci.BuildImage(origPath, cfg, []oci.LayerData{
		oci.NewLayerData(schemaTar, oci.LayerTypeSchema),
		oci.NewLayerData(pluginTar, oci.LayerTypePlugins),
	}); err != nil {
		t.Fatal(err)
	}

	// 2. 构造数据卷（system.db + dist）
	volPath := filepath.Join(dir, "volume")
	if err := os.MkdirAll(filepath.Join(volPath, "dist"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(volPath, "system.db"), []byte("sqlite-data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(volPath, "dist", "index.html"), []byte("<html>v2</html>"), 0644); err != nil {
		t.Fatal(err)
	}

	// 3. commit
	commitPath := filepath.Join(dir, "commit.pmi")
	if err := BuildCommitImage(origPath, volPath, commitPath); err != nil {
		t.Fatalf("BuildCommitImage: %v", err)
	}

	// 4. 回读验证：schema/plugin/data 层都在
	r, err := oci.OpenImage(commitPath)
	if err != nil {
		t.Fatal(err)
	}
	m := r.Manifest()
	var types []string
	for _, ld := range m.Layers {
		types = append(types, ld.Annotations["pmocker.layer.type"])
	}
	has := func(typ string) bool {
		for _, v := range types {
			if v == typ {
				return true
			}
		}
		return false
	}
	if !has("schema") || !has("plugins") || !has("data") {
		t.Fatalf("层类型不完整: %v", types)
	}
	// data 层内容
	files, err := r.ExtractLayerFiles(oci.LayerTypeData)
	if err != nil {
		t.Fatal(err)
	}
	if string(files["system.db"]) != "sqlite-data" {
		t.Errorf("data 层 system.db 内容不符")
	}
	if _, ok := files["dist/index.html"]; !ok {
		t.Error("data 层缺少 dist/index.html")
	}
}

func createTarFromFiles(files map[string][]byte) ([]byte, error) {
	_, tarBytes, err := oci.CreateLayerFromFiles(files, oci.LayerTypeSchema)
	return tarBytes, err
}
