package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
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

// Create 新建配置包（默认 draft，version=1）
func (s *ConfigPackageService) Create(ctx context.Context, pkg pmocker.PMConfigPackage) error {
	if pkg.Status == "" {
		pkg.Status = "draft"
	}
	if pkg.Version == 0 {
		pkg.Version = 1
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

// UpdateSeed 更新 seed_yaml（draft/reviewing 可编辑）
func (s *ConfigPackageService) UpdateSeed(ctx context.Context, id uint, seedYAML string) error {
	var pkg pmocker.PMConfigPackage
	if err := global.GVA_DB.WithContext(ctx).First(&pkg, id).Error; err != nil {
		return err
	}
	if pkg.Status != "draft" && pkg.Status != "reviewing" {
		return errors.New("仅 draft/reviewing 状态可编辑，请先归档或复制为草稿")
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
	copy.Version = 1
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
