package service

import (
	"context"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"gorm.io/gorm"
)

// ConfigVersionService 配置包版本管理（快照 + 回滚）
type ConfigVersionService struct{}

// Snapshot 发布时生成版本快照（在事务内调用）
func (s *ConfigVersionService) Snapshot(db *gorm.DB, pkg *pmocker.PMConfigPackage, version int) error {
	if db == nil {
		db = global.GVA_DB
	}
	return db.Create(&pmocker.PMConfigVersion{
		PackageID:    pkg.ID,
		Version:      version,
		SnapshotYAML: pkg.SeedYAML,
		Flag:         0,
	}).Error
}

// ListVersions 版本历史（按时间倒序）
func (s *ConfigVersionService) ListVersions(ctx context.Context, packageID uint) ([]pmocker.PMConfigVersion, error) {
	var list []pmocker.PMConfigVersion
	err := global.GVA_DB.WithContext(ctx).Where("package_id = ?", packageID).Order("id DESC").Find(&list).Error
	return list, err
}

// Rollback 回滚到指定版本：恢复 snapshot_yaml 到 seed_yaml，重新发布同步，记录回滚标记
func (s *ConfigVersionService) Rollback(ctx context.Context, packageID uint, versionID uint) error {
	db := global.GVA_DB.WithContext(ctx)
	var pkg pmocker.PMConfigPackage
	if err := db.First(&pkg, packageID).Error; err != nil {
		return err
	}
	var ver pmocker.PMConfigVersion
	if err := db.First(&ver, versionID).Error; err != nil {
		return err
	}
	if ver.PackageID != packageID {
		return fmt.Errorf("版本不属于该配置包")
	}
	seed, err := ParseSeedYAML([]byte(ver.SnapshotYAML))
	if err != nil {
		return fmt.Errorf("回滚快照解析失败: %w", err)
	}
	return db.Transaction(func(tx *gorm.DB) error {
		// 恢复 seed_yaml
		if err := tx.Model(&pkg).Update("seed_yaml", ver.SnapshotYAML).Error; err != nil {
			return err
		}
		// 重新发布同步
		syncSvc := &SeedSyncService{}
		if err := syncSvc.Sync(ctx, tx, seed); err != nil {
			return err
		}
		// 状态回 published + 版本递增
		if err := tx.Model(&pkg).Update("status", "published").Error; err != nil {
			return err
		}
		// 记录回滚版本
		return tx.Create(&pmocker.PMConfigVersion{
			PackageID:    pkg.ID,
			Version:      pkg.Version + 1,
			SnapshotYAML: ver.SnapshotYAML,
			Flag:         1,
		}).Error
	})
}
