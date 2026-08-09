package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"gorm.io/gorm"
)

// SeedSyncService 配置包发布时同步 seed_yaml 到运行表
type SeedSyncService struct{}

// Sync 发布同步：seed_yaml → pm_entity_types/pm_field_defs/pm_state_defs/pm_entities/pm_attrs
func (s *SeedSyncService) Sync(ctx context.Context, db *gorm.DB, seed *ConfigPackageSeed) error {
	if err := s.SyncEntityType(ctx, db, seed); err != nil {
		return err
	}
	if err := s.SyncFields(ctx, db, seed); err != nil {
		return err
	}
	if err := s.SyncStates(ctx, db, seed); err != nil {
		return err
	}
	return s.SyncProjectEntities(ctx, db, seed)
}

// SyncEntityType 同步实体类型（upsert，published）
func (s *SeedSyncService) SyncEntityType(ctx context.Context, db *gorm.DB, seed *ConfigPackageSeed) error {
	et := pmocker.PMEntityType{
		TypeCode:   seed.EntityType,
		ModuleCode: seed.Module,
		Name:       seed.Name,
		Status:     "published",
	}
	return db.Where(pmocker.PMEntityType{TypeCode: seed.EntityType}).
		Assign(et).FirstOrCreate(&pmocker.PMEntityType{TypeCode: seed.EntityType}).Error
}

// SyncFields 同步字段定义（upsert，published）
func (s *SeedSyncService) SyncFields(ctx context.Context, db *gorm.DB, seed *ConfigPackageSeed) error {
	for _, f := range seed.Fields {
		optionsJSON := ""
		if len(f.Options) > 0 {
			b, err := json.Marshal(f.Options)
			if err != nil {
				return err
			}
			optionsJSON = string(b)
		}
		fd := pmocker.PMFieldDef{
			EntityType:   seed.EntityType,
			FieldKey:     f.Key,
			FieldLabel:   f.Label,
			DataType:     f.DataType,
			OptionsJSON:  optionsJSON,
			DefaultValue: f.Default,
			Status:       "published",
		}
		if err := db.Where(pmocker.PMFieldDef{EntityType: seed.EntityType, FieldKey: f.Key}).
			Assign(fd).FirstOrCreate(&pmocker.PMFieldDef{EntityType: seed.EntityType, FieldKey: f.Key}).Error; err != nil {
			return err
		}
	}
	return nil
}

// SyncStates 同步状态定义 + 流转规则到 pm_state_defs（upsert）
func (s *SeedSyncService) SyncStates(ctx context.Context, db *gorm.DB, seed *ConfigPackageSeed) error {
	for _, st := range seed.States {
		var actions []map[string]interface{}
		for _, tr := range seed.Transitions {
			if tr.From != st.Status {
				continue
			}
			actions = append(actions, map[string]interface{}{
				"label":    tr.Action,
				"target":   tr.To,
				"action":   tr.Action,
				"rollback": tr.Rollback,
			})
		}
		actionsJSON := "[]"
		if len(actions) > 0 {
			b, err := json.Marshal(actions)
			if err != nil {
				return err
			}
			actionsJSON = string(b)
		}
		sd := pmocker.PMStateDef{
			EntityType:   seed.EntityType,
			Status:       st.Status,
			Label:        st.Label,
			TagType:      st.TagType,
			ActionsJSON:  actionsJSON,
			ConfigStatus: "published",
		}
		if err := db.Where("entity_type = ? AND status = ?", seed.EntityType, st.Status).
			Assign(sd).FirstOrCreate(&pmocker.PMStateDef{EntityType: seed.EntityType, Status: st.Status}).Error; err != nil {
			return err
		}
	}
	return nil
}

// SyncProjectEntities 同步项目实体种子（pm_entities + pm_attrs）
// EPS 配置包的 projects 为树层级（仅建 EPS 节点）；业务配置包通过 project_id/project_code + entities 建业务实体。
func (s *SeedSyncService) SyncProjectEntities(ctx context.Context, db *gorm.DB, seed *ConfigPackageSeed) error {
	for _, p := range seed.Projects {
		if seed.EntityType == "eps_node" {
			if err := s.syncEPSTree(ctx, db, &p, 0); err != nil {
				return err
			}
			continue
		}
		// 解析项目 ID：优先 project_code（查 eps_node 的 code attr），回退 project_id
		projectID := p.ProjectID
		if p.ProjectCode != "" {
			pid, err := resolveProjectID(db, p.ProjectCode)
			if err != nil {
				return err
			}
			projectID = pid
		}
		for _, entities := range p.Entities {
			for _, ent := range entities {
				if err := s.syncEntity(ctx, db, seed.EntityType, projectID, ent); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// resolveProjectID 按项目编码查询 EPS 项目实体 ID
func resolveProjectID(db *gorm.DB, code string) (uint, error) {
	var node pmocker.PMEntity
	err := db.Table("pm_entities").Joins("JOIN pm_attrs ON pm_attrs.entity_id = pm_entities.id").
		Where("pm_entities.entity_type = ? AND pm_attrs.field_key = ? AND pm_attrs.val_string = ?", "eps_node", "code", code).
		First(&node).Error
	if err != nil {
		return 0, fmt.Errorf("项目编码 %s 未找到对应 EPS 项目: %w", code, err)
	}
	return node.ID, nil
}

// syncEPSTree 递归创建 EPS 树节点
func (s *SeedSyncService) syncEPSTree(ctx context.Context, db *gorm.DB, p *ProjectSeed, parentID uint) error {
	title := p.Name
	attrs := map[string]interface{}{
		"code": p.Code,
		"type": p.Type,
	}
	var existing pmocker.PMEntity
	err := db.Where("entity_type = ? AND title = ?", "eps_node", title).First(&existing).Error
	if err == nil {
		for i := range p.Children {
			if err := s.syncEPSTree(ctx, db, &p.Children[i], existing.ID); err != nil {
				return err
			}
		}
		return nil
	}
	node := pmocker.PMEntity{
		ProjectID:  0,
		EntityType: "eps_node",
		Title:      title,
		Status:     p.Status,
		Priority:   p.Priority,
	}
	if parentID > 0 {
		node.ParentID = &parentID
	}
	if err := db.Create(&node).Error; err != nil {
		return err
	}
	for k, v := range attrs {
		if err := setAttr(db, node.ID, k, v); err != nil {
			return err
		}
	}
	for i := range p.Children {
		if err := s.syncEPSTree(ctx, db, &p.Children[i], node.ID); err != nil {
			return err
		}
	}
	return nil
}

// syncEntity 创建/更新单个业务实体（含 attrs），按 entity_type+project_id+title 幂等
func (s *SeedSyncService) syncEntity(ctx context.Context, db *gorm.DB, entityType string, projectID uint, ent map[string]interface{}) error {
	title, _ := ent["title"].(string)
	if title == "" {
		return fmt.Errorf("实体缺少 title: %+v", ent)
	}
	status, _ := ent["status"].(string)
	if status == "" {
		status = "draft"
	}
	var existing pmocker.PMEntity
	err := db.Where("entity_type = ? AND project_id = ? AND title = ?", entityType, projectID, title).First(&existing).Error
	if err == nil {
		return db.Model(&existing).Update("status", status).Error
	}
	node := pmocker.PMEntity{
		ProjectID:  projectID,
		EntityType: entityType,
		Title:      title,
		Status:     status,
	}
	if err := db.Create(&node).Error; err != nil {
		return err
	}
	for k, v := range ent {
		if k == "title" || k == "status" {
			continue
		}
		if err := setAttr(db, node.ID, k, v); err != nil {
			return err
		}
	}
	return nil
}

// setAttr 写入 EAV attr（upsert，按类型）
func setAttr(db *gorm.DB, entityID uint, key string, val interface{}) error {
	var attr pmocker.PMAttr
	switch v := val.(type) {
	case string:
		attr = pmocker.PMAttr{ValString: &v}
	case int:
		i := int64(v)
		attr = pmocker.PMAttr{ValInt: &i}
	case int64:
		attr = pmocker.PMAttr{ValInt: &v}
	case float64:
		if v == float64(int64(v)) {
			i := int64(v)
			attr = pmocker.PMAttr{ValInt: &i}
		} else {
			attr = pmocker.PMAttr{ValDecimal: &v}
		}
	case bool:
		attr = pmocker.PMAttr{ValBool: &v}
	default:
		s := fmt.Sprintf("%v", v)
		attr = pmocker.PMAttr{ValString: &s}
	}
	return db.Where(pmocker.PMAttr{EntityID: entityID, FieldKey: key}).
		Assign(attr).FirstOrCreate(&pmocker.PMAttr{EntityID: entityID, FieldKey: key}).Error
}
