package oci

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CreateLayer 从目录创建 tar 层，返回描述符和 tar 字节
func CreateLayer(srcDir string, layerType LayerType) (*LayerDescriptor, []byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		hdr := &tar.Header{
			Name: rel,
			Mode: 0644,
			Size: int64(len(data)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		_, err = tw.Write(data)
		return err
	})
	if err != nil {
		tw.Close()
		return nil, nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, nil, err
	}
	tarBytes := buf.Bytes()
	return descriptorFor(tarBytes, layerType), tarBytes, nil
}

// CreateLayerFromFiles 从文件映射创建 tar 层
func CreateLayerFromFiles(files map[string][]byte, layerType LayerType) (*LayerDescriptor, []byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for name, data := range files {
		name = strings.TrimPrefix(name, "/")
		hdr := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(data)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			tw.Close()
			return nil, nil, err
		}
		if _, err := tw.Write(data); err != nil {
			tw.Close()
			return nil, nil, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, nil, err
	}
	tarBytes := buf.Bytes()
	return descriptorFor(tarBytes, layerType), tarBytes, nil
}

func descriptorFor(data []byte, layerType LayerType) *LayerDescriptor {
	h := sha256.Sum256(data)
	digest := DigestPrefix + fmt.Sprintf("%x", h)
	return &LayerDescriptor{
		Descriptor: Descriptor{
			MediaType: MediaTypeLayerPrefix + string(layerType) + ".v1.tar",
			Digest:    digest,
			Size:      int64(len(data)),
		},
		Annotations: map[string]string{
			"pmocker.layer.type": string(layerType),
		},
	}
}
