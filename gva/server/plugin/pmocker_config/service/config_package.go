package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"gorm.io/gorm"
)

type ConfigPackageService struct{}

// List 配置包列表；includeDraft=true 返回全部
func (s *ConfigPackageService) List(ctx context.Context, includeDraft bool) ([]pmocker.PMConfigPackage, error) {
	var list []pmocker.PMConfigPackage
	db := global.GVA_DB.WithContext(ctx)
	if !includeDraft {
		db = db.Where("status = ?", "published")
	}
	err := db.Order("id").Find(&list).Error
	return list, err
}

// Create 新建配置包（默认 draft，version=0 未发布）
func (s *ConfigPackageService) Create(ctx context.Context, pkg pmocker.PMConfigPackage) error {
	if pkg.Status == "" {
		pkg.Status = "draft"
	}
	return global.GVA_DB.WithContext(ctx).Create(&pkg).Error
}

// Get 获取配置包详情
func (s *ConfigPackageService) Get(ctx context.Context, id uint) (*pmocker.PMConfigPackage, error) {
	var pkg pmocker.PMConfigPackage
	if err := global.GVA_DB.WithContext(ctx).First(&pkg, id).Error; err != nil {
		return nil, err
	}
	return &pkg, nil
}

// UpdateSeed 更新 seed_yaml（draft/reviewing/published 可编辑；archived 需先恢复）
func (s *ConfigPackageService) UpdateSeed(ctx context.Context, id uint, seedYAML string) error {
	var pkg pmocker.PMConfigPackage
	if err := global.GVA_DB.WithContext(ctx).First(&pkg, id).Error; err != nil {
		return err
	}
	if pkg.Status == "archived" {
		return errors.New("archived 状态不可编辑，请先恢复为草稿")
	}
	return global.GVA_DB.WithContext(ctx).Model(&pkg).Update("seed_yaml", seedYAML).Error
}

// CopyAsDraft 复制为 draft：code 加 -copy，seed_yaml 原样复制
func (s *ConfigPackageService) CopyAsDraft(ctx context.Context, id uint) error {
	var src pmocker.PMConfigPackage
	if err := global.GVA_DB.WithContext(ctx).First(&src, id).Error; err != nil {
		return err
	}
	copy := src
	copy.ID = 0
	copy.Code = src.Code + "-copy"
	copy.Name = src.Name + "(副本)"
	copy.Version = 0
	copy.Status = "draft"
	return global.GVA_DB.WithContext(ctx).Create(&copy).Error
}

// Delete 删除（仅 draft）
func (s *ConfigPackageService) Delete(ctx context.Context, id uint) error {
	var pkg pmocker.PMConfigPackage
	if err := global.GVA_DB.WithContext(ctx).First(&pkg, id).Error; err != nil {
		return err
	}
	if pkg.Status != "draft" {
		return fmt.Errorf("仅 draft 状态可删除（当前 %s），请先归档", pkg.Status)
	}
	return global.GVA_DB.WithContext(ctx).Delete(&pkg).Error
}

// Publish 发布配置包：状态机校验 + 解析 seed_yaml + 同步运行表 + 版本号递增 + 版本快照
// draft/reviewing/published → published（published 重新发布 = 新版本）
func (s *ConfigPackageService) Publish(ctx context.Context, id uint) error {
	db := global.GVA_DB.WithContext(ctx)
	var pkg pmocker.PMConfigPackage
	if err := db.First(&pkg, id).Error; err != nil {
		return err
	}
	if pkg.Status == "archived" {
		return fmt.Errorf("archived 配置包不可发布，请先恢复")
	}
	seed, err := ParseSeedYAML([]byte(pkg.SeedYAML))
	if err != nil {
		return fmt.Errorf("seed_yaml 解析失败: %w", err)
	}
	// 事务：同步运行表 + 状态流转 + 版本递增 + 版本快照
	return db.Transaction(func(tx *gorm.DB) error {
		syncSvc := &SeedSyncService{}
		if err := syncSvc.Sync(ctx, tx, seed); err != nil {
			return fmt.Errorf("同步运行表失败: %w", err)
		}
		if err := tx.Model(&pkg).Update("status", "published").Error; err != nil {
			return err
		}
		// 版本递增并记录快照（快照版本 = 递增后的版本号）
		newVersion := pkg.Version + 1
		if err := tx.Model(&pkg).Update("version", newVersion).Error; err != nil {
			return err
		}
		verSvc := &ConfigVersionService{}
		return verSvc.Snapshot(tx, &pkg, newVersion)
	})
}

// Archive 归档配置包（published → archived）
func (s *ConfigPackageService) Archive(ctx context.Context, id uint) error {
	var pkg pmocker.PMConfigPackage
	if err := global.GVA_DB.WithContext(ctx).First(&pkg, id).Error; err != nil {
		return err
	}
	if pkg.Status != "published" {
		return fmt.Errorf("仅 published 可归档（当前 %s）", pkg.Status)
	}
	return global.GVA_DB.WithContext(ctx).Model(&pkg).Update("status", "archived").Error
}

// Restore 恢复为草稿（archived → draft）
func (s *ConfigPackageService) Restore(ctx context.Context, id uint) error {
	var pkg pmocker.PMConfigPackage
	if err := global.GVA_DB.WithContext(ctx).First(&pkg, id).Error; err != nil {
		return err
	}
	if pkg.Status != "archived" {
		return fmt.Errorf("仅 archived 可恢复（当前 %s）", pkg.Status)
	}
	return global.GVA_DB.WithContext(ctx).Model(&pkg).Update("status", "draft").Error
}

// ListStateDefsPublic 已发布状态流转定义（前端 statusTransitions 读取）
func (s *ConfigPackageService) ListStateDefsPublic(ctx context.Context, entityType string) ([]pmocker.PMStateDef, error) {
	var list []pmocker.PMStateDef
	db := global.GVA_DB.WithContext(ctx).Where("config_status = ?", "published")
	if entityType != "" {
		db = db.Where("entity_type = ?", entityType)
	}
	err := db.Order("entity_type, sort").Find(&list).Error
	return list, err
}

// LoadSeedPackage 灌入单个配置包 seed_yaml 到 pm_config_packages（幂等：code 已存在跳过）
func (s *ConfigPackageService) LoadSeedPackage(ctx context.Context, code string, seedYAML string) error {
	var count int64
	global.GVA_DB.WithContext(ctx).Model(&pmocker.PMConfigPackage{}).Where("code = ?", code).Count(&count)
	if count > 0 {
		return nil
	}
	// 从 seed_yaml 解析基本信息
	seed, err := ParseSeedYAML([]byte(seedYAML))
	if err != nil {
		return err
	}
	pkg := pmocker.PMConfigPackage{
		Code:       code,
		Name:       seed.Name,
		EntityType: seed.EntityType,
		Module:     seed.Module,
		Status:     "draft",
		SeedYAML:   seedYAML,
	}
	return global.GVA_DB.WithContext(ctx).Create(&pkg).Error
}

