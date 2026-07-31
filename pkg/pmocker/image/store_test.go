package image

import (
	"path/filepath"
	"testing"

	"github.com/pmocker-io/pmocker/pkg/pmocker/oci"
)

func TestStoreAddAndList(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	// 构建测试镜像
	pmiPath := filepath.Join(tmpDir, "test.pmi")
	cfg := oci.NewConfig("Test", "1.0.0", []string{"requirement"})
	schemaFiles := map[string][]byte{
		"requirement.yaml": []byte("entity_type: requirement\nfields: []"),
	}
	_, schemaTar, _ := oci.CreateLayerFromFiles(schemaFiles, oci.LayerTypeSchema)
	schemaLayer := oci.NewLayerData(schemaTar, oci.LayerTypeSchema)
	oci.BuildImage(pmiPath, cfg, []oci.LayerData{schemaLayer})

	// 添加镜像
	info, err := store.AddImage(pmiPath, "test", "latest")
	if err != nil {
		t.Fatalf("AddImage: %v", err)
	}
	if info.Digest == "" {
		t.Error("digest is empty")
	}

	// 列出
	images, err := store.ListImages()
	if err != nil {
		t.Fatalf("ListImages: %v", err)
	}
	if len(images) != 1 {
		t.Fatalf("image count = %d, want 1", len(images))
	}
	if images[0].Name != "test" || images[0].Tag != "latest" {
		t.Errorf("name/tag = %s/%s", images[0].Name, images[0].Tag)
	}

	// ResolveImage
	resolved, err := store.ResolveImage("test", "latest")
	if err != nil {
		t.Fatalf("ResolveImage: %v", err)
	}
	if resolved.Digest != info.Digest {
		t.Errorf("digest mismatch")
	}

	// RemoveImage
	if err := store.RemoveImage(info.Digest); err != nil {
		t.Fatalf("RemoveImage: %v", err)
	}
	images, _ = store.ListImages()
	if len(images) != 0 {
		t.Errorf("after remove, count = %d", len(images))
	}
}
