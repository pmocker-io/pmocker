package oci

import (
	"archive/tar"
	"bytes"
	"testing"
)

func TestCreateLayerFromFiles(t *testing.T) {
	files := map[string][]byte{
		"schema.yaml": []byte("entity_type: test\nfields: []"),
		"seed.yaml":   []byte("dictionaries: []"),
	}
	desc, data, err := CreateLayerFromFiles(files, LayerTypeSchema)
	if err != nil {
		t.Fatalf("CreateLayerFromFiles: %v", err)
	}
	if desc.Digest == "" {
		t.Error("digest is empty")
	}
	if desc.Size != int64(len(data)) {
		t.Errorf("size = %d, want %d", desc.Size, len(data))
	}
	if desc.Annotations["pmocker.layer.type"] != "schema" {
		t.Errorf("layer type = %s", desc.Annotations["pmocker.layer.type"])
	}
	// 验证 tar 内容
	tr := tar.NewReader(bytes.NewReader(data))
	count := 0
	for {
		hdr, err := tr.Next()
		if err != nil {
			break
		}
		count++
		if hdr.Name != "schema.yaml" && hdr.Name != "seed.yaml" {
			t.Errorf("unexpected file: %s", hdr.Name)
		}
	}
	if count != 2 {
		t.Errorf("file count = %d, want 2", count)
	}
}

func TestCreateLayerDeterministic(t *testing.T) {
	files := map[string][]byte{
		"a.yaml": []byte("hello"),
	}
	desc1, _, _ := CreateLayerFromFiles(files, LayerTypeSchema)
	desc2, _, _ := CreateLayerFromFiles(files, LayerTypeSchema)
	if desc1.Digest != desc2.Digest {
		t.Errorf("digests differ: %s != %s", desc1.Digest, desc2.Digest)
	}
}
