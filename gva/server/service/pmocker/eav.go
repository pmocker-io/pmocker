package pmocker

import (
	"context"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
)

// EAVService 实现 eav.EAVStore 接口
type EAVService struct{}

// LoadEntityType 加载实体类型
func (s *EAVService) LoadEntityType(ctx context.Context, typeCode string) (*eavtypes.EntityType, error) {
	var et pmocker.PMEntityType
	if err := global.GVA_DB.WithContext(ctx).Where("type_code = ?", typeCode).First(&et).Error; err != nil {
		return nil, err
	}
	return &eavtypes.EntityType{
		TypeCode:   et.TypeCode,
		ModuleCode: et.ModuleCode,
		Name:       et.Name,
		Icon:       et.Icon,
		IconColor:  et.IconColor,
	}, nil
}

// LoadFieldDefs 加载字段定义
func (s *EAVService) LoadFieldDefs(ctx context.Context, typeCode string) ([]eavtypes.FieldDef, error) {
	var defs []pmocker.PMFieldDef
	if err := global.GVA_DB.WithContext(ctx).Where("entity_type = ?", typeCode).Find(&defs).Error; err != nil {
		return nil, err
	}
	result := make([]eavtypes.FieldDef, len(defs))
	for i, d := range defs {
		result[i] = eavtypes.FieldDef{
			EntityType:   d.EntityType,
			FieldKey:     d.FieldKey,
			FieldLabel:   d.FieldLabel,
			DataType:     eavtypes.DataType(d.DataType),
			OptionsJSON:  d.OptionsJSON,
			DefaultValue: d.DefaultValue,
			Validators:   d.Validators,
		}
	}
	return result, nil
}

// RegisterEntityType 注册实体类型（upsert）
func (s *EAVService) RegisterEntityType(ctx context.Context, et eavtypes.EntityType) error {
	return global.GVA_DB.WithContext(ctx).
		Where(pmocker.PMEntityType{TypeCode: et.TypeCode}).
		Assign(pmocker.PMEntityType{
			ModuleCode: et.ModuleCode,
			Name:       et.Name,
			Icon:       et.Icon,
			IconColor:  et.IconColor,
		}).FirstOrCreate(&pmocker.PMEntityType{TypeCode: et.TypeCode}).Error
}

// RegisterFieldDef 注册字段定义（upsert）
func (s *EAVService) RegisterFieldDef(ctx context.Context, fd eavtypes.FieldDef) error {
	return global.GVA_DB.WithContext(ctx).
		Where(pmocker.PMFieldDef{EntityType: fd.EntityType, FieldKey: fd.FieldKey}).
		Assign(pmocker.PMFieldDef{
			FieldLabel:   fd.FieldLabel,
			DataType:     string(fd.DataType),
			OptionsJSON:  fd.OptionsJSON,
			DefaultValue: fd.DefaultValue,
			Validators:   fd.Validators,
		}).FirstOrCreate(&pmocker.PMFieldDef{
			EntityType: fd.EntityType,
			FieldKey:   fd.FieldKey,
		}).Error
}

// CreateEntity 创建实体（含动态字段）
func (s *EAVService) CreateEntity(ctx context.Context, e eavtypes.Entity) (uint, error) {
	db := global.GVA_DB.WithContext(ctx)
	entity := pmocker.PMEntity{
		ProjectID:  e.ProjectID,
		EntityType: e.EntityType,
		ParentID:   e.ParentID,
		Title:      e.Title,
		Status:     e.Status,
		OwnerID:    e.OwnerID,
		BaselineID: e.BaselineID,
		Seq:        e.Seq,
		CreatedBy:  e.CreatedBy,
	}
	if err := db.Create(&entity).Error; err != nil {
		return 0, err
	}
	for key, val := range e.Attrs {
		if err := s.setAttr(ctx, entity.ID, key, val); err != nil {
			return 0, fmt.Errorf("set attr %s: %w", key, err)
		}
	}
	return entity.ID, nil
}

// GetEntity 获取实体（含动态字段）
func (s *EAVService) GetEntity(ctx context.Context, id uint) (*eavtypes.Entity, error) {
	var entity pmocker.PMEntity
	if err := global.GVA_DB.WithContext(ctx).First(&entity, id).Error; err != nil {
		return nil, err
	}
	var attrs []pmocker.PMAttr
	if err := global.GVA_DB.WithContext(ctx).Where("entity_id = ?", id).Find(&attrs).Error; err != nil {
		return nil, err
	}
	attrMap := make(map[string]interface{})
	for _, a := range attrs {
		attrMap[a.FieldKey] = s.readAttrValue(a)
	}
	return &eavtypes.Entity{
		ID:         entity.ID,
		ProjectID:  entity.ProjectID,
		EntityType: entity.EntityType,
		ParentID:   entity.ParentID,
		Title:      entity.Title,
		Status:     entity.Status,
		OwnerID:    entity.OwnerID,
		BaselineID: entity.BaselineID,
		Seq:        entity.Seq,
		Attrs:      attrMap,
		CreatedAt:  entity.CreatedAt,
		UpdatedAt:  entity.UpdatedAt,
	}, nil
}

// UpdateEntity 更新实体
func (s *EAVService) UpdateEntity(ctx context.Context, e eavtypes.Entity) error {
	db := global.GVA_DB.WithContext(ctx)
	if err := db.Model(&pmocker.PMEntity{}).Where("id = ?", e.ID).Updates(map[string]interface{}{
		"title":    e.Title,
		"status":   e.Status,
		"owner_id": e.OwnerID,
		"seq":      e.Seq,
	}).Error; err != nil {
		return err
	}
	for key, val := range e.Attrs {
		if err := s.setAttr(ctx, e.ID, key, val); err != nil {
			return fmt.Errorf("set attr %s: %w", key, err)
		}
	}
	return nil
}

// DeleteEntity 删除实体（软删除）
func (s *EAVService) DeleteEntity(ctx context.Context, id uint) error {
	return global.GVA_DB.WithContext(ctx).Delete(&pmocker.PMEntity{}, id).Error
}

// ListEntities 列出实体
func (s *EAVService) ListEntities(ctx context.Context, projectID uint, typeCode string, offset, limit int) ([]eavtypes.Entity, int64, error) {
	db := global.GVA_DB.WithContext(ctx).Model(&pmocker.PMEntity{}).
		Where("project_id = ? AND entity_type = ?", projectID, typeCode)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var entities []pmocker.PMEntity
	if err := db.Offset(offset).Limit(limit).Order("seq, id").Find(&entities).Error; err != nil {
		return nil, 0, err
	}
	// 批量加载属性，按 entityID 分组
	entityIDs := make([]uint, len(entities))
	for i, e := range entities {
		entityIDs[i] = e.ID
	}
	attrMap := make(map[uint]map[string]interface{})
	if len(entityIDs) > 0 {
		var attrs []pmocker.PMAttr
		if err := global.GVA_DB.WithContext(ctx).Where("entity_id IN ?", entityIDs).Find(&attrs).Error; err != nil {
			return nil, 0, err
		}
		for _, a := range attrs {
			if attrMap[a.EntityID] == nil {
				attrMap[a.EntityID] = make(map[string]interface{})
			}
			if a.ValString != nil {
				attrMap[a.EntityID][a.FieldKey] = *a.ValString
			} else if a.ValInt != nil {
				attrMap[a.EntityID][a.FieldKey] = *a.ValInt
			} else if a.ValDecimal != nil {
				attrMap[a.EntityID][a.FieldKey] = *a.ValDecimal
			} else if a.ValBool != nil {
				attrMap[a.EntityID][a.FieldKey] = *a.ValBool
			} else if a.ValDate != nil {
				attrMap[a.EntityID][a.FieldKey] = *a.ValDate
			} else if a.ValDateTime != nil {
				attrMap[a.EntityID][a.FieldKey] = *a.ValDateTime
			} else if a.ValJSON != nil {
				attrMap[a.EntityID][a.FieldKey] = *a.ValJSON
			} else if a.ValRef != nil {
				attrMap[a.EntityID][a.FieldKey] = *a.ValRef
			}
		}
	}
	result := make([]eavtypes.Entity, len(entities))
	for i, e := range entities {
		result[i] = eavtypes.Entity{
			ID:         e.ID,
			ProjectID:  e.ProjectID,
			EntityType: e.EntityType,
			ParentID:   e.ParentID,
			Title:      e.Title,
			Status:     e.Status,
			OwnerID:    e.OwnerID,
			Seq:        e.Seq,
			CreatedAt:  e.CreatedAt,
			UpdatedAt:  e.UpdatedAt,
			Attrs:      attrMap[e.ID],
		}
	}
	return result, total, nil
}

// AddRelation 添加关系
func (s *EAVService) AddRelation(ctx context.Context, srcID, dstID uint, relType string) error {
	return global.GVA_DB.WithContext(ctx).
		Where(pmocker.PMRelation{SrcID: srcID, DstID: dstID, RelationType: relType}).
		FirstOrCreate(&pmocker.PMRelation{SrcID: srcID, DstID: dstID, RelationType: relType}).Error
}

// ListRelations 列出关系
func (s *EAVService) ListRelations(ctx context.Context, entityID uint) ([]eavtypes.Relation, error) {
	var rels []pmocker.PMRelation
	if err := global.GVA_DB.WithContext(ctx).Where("src_id = ? OR dst_id = ?", entityID, entityID).Find(&rels).Error; err != nil {
		return nil, err
	}
	result := make([]eavtypes.Relation, len(rels))
	for i, r := range rels {
		result[i] = eavtypes.Relation{
			ID:           r.ID,
			SrcID:        r.SrcID,
			DstID:        r.DstID,
			RelationType: r.RelationType,
		}
	}
	return result, nil
}

// setAttr 设置属性值（upsert）
func (s *EAVService) setAttr(ctx context.Context, entityID uint, key string, val interface{}) error {
	var attr pmocker.PMAttr
	s.writeAttrValue(&attr, val)
	return global.GVA_DB.WithContext(ctx).
		Where(pmocker.PMAttr{EntityID: entityID, FieldKey: key}).
		Assign(attr).
		FirstOrCreate(&pmocker.PMAttr{EntityID: entityID, FieldKey: key}).Error
}

// writeAttrValue 根据 Go 类型写入对应列
func (s *EAVService) writeAttrValue(attr *pmocker.PMAttr, val interface{}) {
	switch v := val.(type) {
	case string:
		attr.ValString = &v
	case int:
		i := int64(v)
		attr.ValInt = &i
	case int64:
		attr.ValInt = &v
	case float64:
		attr.ValDecimal = &v
	case bool:
		attr.ValBool = &v
	default:
		s := fmt.Sprintf("%v", val)
		attr.ValString = &s
	}
}

// readAttrValue 读取属性值
func (s *EAVService) readAttrValue(attr pmocker.PMAttr) interface{} {
	if attr.ValString != nil {
		return *attr.ValString
	}
	if attr.ValInt != nil {
		return *attr.ValInt
	}
	if attr.ValDecimal != nil {
		return *attr.ValDecimal
	}
	if attr.ValBool != nil {
		return *attr.ValBool
	}
	if attr.ValJSON != nil {
		return *attr.ValJSON
	}
	return nil
}

// 确保 EAVService 实现了 EAVStore 接口
var _ eavtypes.EAVStore = (*EAVService)(nil)
