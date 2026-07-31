package oci

import (
	"archive/tar"
	"fmt"
	"os"
)

// LayerData 持有层的原始数据
type LayerData struct {
	Type     LayerType
	TarBytes []byte
	Desc     *LayerDescriptor
}

// NewLayerData 从 tar 字节创建 LayerData
func NewLayerData(tarBytes []byte, layerType LayerType) LayerData {
	desc := descriptorFor(tarBytes, layerType)
	return LayerData{Type: layerType, TarBytes: tarBytes, Desc: desc}
}

// BuildImage 组装 .pmi 文件到指定路径
func BuildImage(outPath string, cfg Config, layers []LayerData) error {
	// 序列化 config
	cfgBytes, err := JSONMarshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	// 构建 manifest
	var descs []LayerDescriptor
	for _, l := range layers {
		descs = append(descs, *l.Desc)
	}
	cfgDesc := descriptorFor(cfgBytes, LayerType(""))
	manifest := NewManifest(cfg, cfgDesc.Digest, cfgDesc.Size, descs)
	// 设置镜像 annotations
	if manifest.Annotations == nil {
		manifest.Annotations = map[string]string{}
	}
	manifest.Annotations["pmocker.image.title"] = cfg.Methodology
	manifest.Annotations["pmocker.image.version"] = cfg.Version
	manifest.Annotations["pmocker.image.methodology"] = cfg.Methodology

	manBytes, err := JSONMarshal(manifest)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}

	// 写 .pmi tar
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create %s: %w", outPath, err)
	}
	defer f.Close()

	tw := tar.NewWriter(f)
	defer tw.Close()

	// 写 manifest.json
	if err := writeTarFile(tw, "manifest.json", manBytes); err != nil {
		return err
	}
	// 写 config.json
	if err := writeTarFile(tw, "config.json", cfgBytes); err != nil {
		return err
	}
	// 写各层
	for _, l := range layers {
		if err := writeTarFile(tw, l.Desc.Digest+".tar", l.TarBytes); err != nil {
			return err
		}
	}
	return nil
}

func writeTarFile(tw *tar.Writer, name string, data []byte) error {
	hdr := &tar.Header{
		Name: name,
		Mode: 0644,
		Size: int64(len(data)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return fmt.Errorf("write header %s: %w", name, err)
	}
	if _, err := tw.Write(data); err != nil {
		return fmt.Errorf("write data %s: %w", name, err)
	}
	return nil
}
