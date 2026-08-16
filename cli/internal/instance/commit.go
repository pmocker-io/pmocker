package instance

import (
	"fmt"
	"time"

	"github.com/pmocker-io/pmocker/pkg/pmocker/oci"
)

// BuildCommitImage 基于原镜像 + 实例数据卷快照，构建新的 .pmi 镜像。
// 原镜像的 schema/plugin/theme/assets 层原样保留（跳过旧 data 层），
// 追加当前数据卷的 data 层（system.db + dist + uploads）。
// 产出写入 outPath，供 commit/export 使用。
func BuildCommitImage(origPMIPath, volumePath, outPath string) error {
	reader, err := oci.OpenImage(origPMIPath)
	if err != nil {
		return fmt.Errorf("open original image: %w", err)
	}
	cfg := reader.Config()

	// 标记为数据快照镜像（时间戳保证 config digest 唯一，避免与源镜像冲突）
	cfg.Version = cfg.Version + "-data-" + time.Now().Format("20060102T150405")

	// 提取原镜像各层（保留非 data 层）
	var layers []oci.LayerData
	for _, ld := range reader.Manifest().Layers {
		typ := oci.LayerType(ld.Annotations["pmocker.layer.type"])
		if typ == oci.LayerTypeData {
			continue // 旧 data 层丢弃，用新快照替代
		}
		tarBytes, err := reader.ExtractLayer(ld.Digest)
		if err != nil {
			return fmt.Errorf("extract layer %s: %w", ld.Digest, err)
		}
		layers = append(layers, oci.NewLayerData(tarBytes, typ))
	}

	// 数据卷快照 → data 层
	snapshot, err := SnapshotVolume(volumePath, true)
	if err != nil {
		return fmt.Errorf("snapshot volume: %w", err)
	}
	layers = append(layers, oci.NewLayerData(snapshot, oci.LayerTypeData))

	// 构建新镜像
	if err := oci.BuildImage(outPath, cfg, layers); err != nil {
		return fmt.Errorf("build commit image: %w", err)
	}
	return nil
}
