package oci

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ImageReader 读取 .pmi 镜像文件
type ImageReader struct {
	path      string
	manifest  Manifest
	config    Config
	layerData map[string][]byte // digest -> tar bytes
}

// OpenImage 打开 .pmi 文件并解析
func OpenImage(path string) (*ImageReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	tr := tar.NewReader(f)
	r := &ImageReader{path: path, layerData: make(map[string][]byte)}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tar: %w", err)
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", hdr.Name, err)
		}
		switch hdr.Name {
		case "manifest.json":
			if err := json.Unmarshal(data, &r.manifest); err != nil {
				return nil, fmt.Errorf("parse manifest: %w", err)
			}
		case "config.json":
			if err := json.Unmarshal(data, &r.config); err != nil {
				return nil, fmt.Errorf("parse config: %w", err)
			}
		default:
			// layer tar 文件名格式：sha256:xxx.tar
			if len(hdr.Name) > 4 && hdr.Name[len(hdr.Name)-4:] == ".tar" {
				digest := hdr.Name[:len(hdr.Name)-4]
				r.layerData[digest] = data
			}
		}
	}
	if r.manifest.SchemaVersion == 0 {
		return nil, fmt.Errorf("manifest.json not found in %s", path)
	}
	return r, nil
}

// Manifest 返回镜像清单
func (r *ImageReader) Manifest() Manifest { return r.manifest }

// Config 返回镜像配置
func (r *ImageReader) Config() Config { return r.config }

// ExtractLayer 按 digest 提取层的 tar 字节
func (r *ImageReader) ExtractLayer(digest string) ([]byte, error) {
	data, ok := r.layerData[digest]
	if !ok {
		return nil, fmt.Errorf("layer %s not found", digest)
	}
	return data, nil
}

// ExtractLayerByType 按类型提取层
func (r *ImageReader) ExtractLayerByType(layerType LayerType) ([]byte, error) {
	for _, layer := range r.manifest.Layers {
		if layer.Annotations["pmocker.layer.type"] == string(layerType) {
			return r.ExtractLayer(layer.Digest)
		}
	}
	return nil, fmt.Errorf("layer type %s not found", layerType)
}

// ExtractLayerFiles 按类型提取层并返回文件名->内容映射
func (r *ImageReader) ExtractLayerFiles(layerType LayerType) (map[string][]byte, error) {
	data, err := r.ExtractLayerByType(layerType)
	if err != nil {
		return nil, err
	}
	return extractTarFiles(data)
}

// extractTarFiles 从 tar 字节提取所有文件
func extractTarFiles(data []byte) (map[string][]byte, error) {
	files := make(map[string][]byte)
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		content, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		files[hdr.Name] = content
	}
	return files, nil
}
