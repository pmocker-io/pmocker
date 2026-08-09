package service

import (
	"context"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

type ConfigService struct{}

// ListEntityTypes 实体类型列表；includeDraft=true 返回全部，false 仅 published
func (s *ConfigService) ListEntityTypes(ctx context.Context, includeDraft bool) ([]pmocker.PMEntityType, error) {
	var list []pmocker.PMEntityType
	db := global.GVA_DB.WithContext(ctx)
	if !includeDraft {
		db = db.Where("status = ?", "published")
	}
	err := db.Order("id").Find(&list).Error
	return list, err
}

// CreateEntityType 新建实体类型，默认 status=published（创建即生效）
func (s *ConfigService) CreateEntityType(ctx context.Context, et pmocker.PMEntityType) error {
	if et.Status == "" {
		et.Status = "published"
	}
	return global.GVA_DB.WithContext(ctx).Create(&et).Error
}

// CopyAsDraft 复制为 draft：源对象 -> 新对象(type_code 加 -copy)，status=draft
func (s *ConfigService) CopyAsDraft(ctx context.Context, table string, id uint) error {
	db := global.GVA_DB.WithContext(ctx)
	switch table {
	case "pm_entity_types":
		var src pmocker.PMEntityType
		if err := db.First(&src, id).Error; err != nil { return err }
		copy := src
		copy.ID = 0
		copy.TypeCode = src.TypeCode + "-copy"
		copy.Status = "draft"
		return db.Create(&copy).Error
	default:
		return fmt.Errorf("暂不支持复制表: %s", table)
	}
}

// ListFieldDefs 字段定义列表（按实体类型 + status 过滤）
func (s *ConfigService) ListFieldDefs(ctx context.Context, entityType string, includeDraft bool) ([]pmocker.PMFieldDef, error) {
	var list []pmocker.PMFieldDef
	db := global.GVA_DB.WithContext(ctx).Where("entity_type = ?", entityType)
	if !includeDraft {
		db = db.Where("status = ?", "published")
	}
	err := db.Order("id").Find(&list).Error
	return list, err
}

// ListStateDefs 状态流转定义列表
func (s *ConfigService) ListStateDefs(ctx context.Context, entityType string, includeDraft bool) ([]pmocker.PMStateDef, error) {
	var list []pmocker.PMStateDef
	db := global.GVA_DB.WithContext(ctx)
	if entityType != "" {
		db = db.Where("entity_type = ?", entityType)
	}
	if !includeDraft {
		db = db.Where("config_status = ?", "published")
	}
	err := db.Order("entity_type, sort").Find(&list).Error
	return list, err
}

// ListWorkflows 工作流定义列表
func (s *ConfigService) ListWorkflows(ctx context.Context, includeDraft bool) ([]pmocker.PMWorkflowDef, error) {
	var list []pmocker.PMWorkflowDef
	db := global.GVA_DB.WithContext(ctx)
	if !includeDraft {
		db = db.Where("status = ?", "published")
	}
	err := db.Order("id").Find(&list).Error
	return list, err
}

// ListSeedEntities 业务种子实体列表（复用 EAV 实体查询）
func (s *ConfigService) ListSeedEntities(ctx context.Context, projectID uint, entityType string, offset, limit int) ([]pmocker.PMEntity, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&pmocker.PMEntity{}).Where("entity_type = ?", entityType)
	if projectID > 0 {
		db = db.Where("project_id = ?", projectID)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil { return nil, 0, err }
	var list []pmocker.PMEntity
	if err := db.Offset(offset).Limit(limit).Order("id").Find(&list).Error; err != nil { return nil, 0, err }
	return list, total, nil
}
