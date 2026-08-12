package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"gorm.io/gorm"
)

// SeedSyncService 配置包发布时同步 seed_yaml 到运行表
type SeedSyncService struct{}

// Sync 发布同步：遍历所有模块，seed_yaml → pm_entity_types/pm_field_defs/pm_state_defs/pm_entities/pm_attrs
func (s *SeedSyncService) Sync(ctx context.Context, db *gorm.DB, seed *ConfigPackageSeed) error {
	// 模块名排序，保证确定性（EPS 先同步，业务模块后）
	for _, module := range sortedModuleKeys(seed.Modules) {
		ms := seed.Modules[module]
		if err := s.syncModule(ctx, db, module, &ms); err != nil {
			return fmt.Errorf("模块 %s 同步失败: %w", module, err)
		}
	}
	return nil
}

// syncModule 同步单个模块（实体类型/字段/状态/项目种子）
func (s *SeedSyncService) syncModule(ctx context.Context, db *gorm.DB, module string, ms *ModuleSeed) error {
	if err := s.SyncEntityType(ctx, db, module, ms); err != nil {
		return err
	}
	if err := s.SyncFields(ctx, db, ms); err != nil {
		return err
	}
	if err := s.SyncStates(ctx, db, ms); err != nil {
		return err
	}
	return s.SyncProjectEntities(ctx, db, module, ms)
}

// SyncEntityType 同步实体类型（upsert，published）
func (s *SeedSyncService) SyncEntityType(ctx context.Context, db *gorm.DB, module string, ms *ModuleSeed) error {
	et := pmocker.PMEntityType{
		TypeCode:   ms.EntityType,
		ModuleCode: module,
		Name:       ms.Name,
		Status:     "published",
	}
	return db.Where(pmocker.PMEntityType{TypeCode: ms.EntityType}).
		Assign(et).FirstOrCreate(&pmocker.PMEntityType{TypeCode: ms.EntityType}).Error
}

// SyncFields 同步字段定义（upsert，published）
func (s *SeedSyncService) SyncFields(ctx context.Context, db *gorm.DB, ms *ModuleSeed) error {
	for _, f := range ms.Fields {
		optionsJSON := ""
		if len(f.Options) > 0 {
			b, err := json.Marshal(f.Options)
			if err != nil {
				return err
			}
			optionsJSON = string(b)
		}
		fd := pmocker.PMFieldDef{
			EntityType:   ms.EntityType,
			FieldKey:     f.Key,
			FieldLabel:   f.Label,
			DataType:     f.DataType,
			OptionsJSON:  optionsJSON,
			DefaultValue: f.Default,
			Status:       "published",
		}
		if err := db.Where(pmocker.PMFieldDef{EntityType: ms.EntityType, FieldKey: f.Key}).
			Assign(fd).FirstOrCreate(&pmocker.PMFieldDef{EntityType: ms.EntityType, FieldKey: f.Key}).Error; err != nil {
			return err
		}
	}
	return nil
}

// SyncStates 同步状态定义 + 流转规则到 pm_state_defs（upsert）
func (s *SeedSyncService) SyncStates(ctx context.Context, db *gorm.DB, ms *ModuleSeed) error {
	for _, st := range ms.States {
		var actions []map[string]interface{}
		for _, tr := range ms.Transitions {
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
			EntityType:   ms.EntityType,
			Status:       st.Status,
			Label:        st.Label,
			TagType:      st.TagType,
			ActionsJSON:  actionsJSON,
			ConfigStatus: "published",
		}
		if err := db.Where("entity_type = ? AND status = ?", ms.EntityType, st.Status).
			Assign(sd).FirstOrCreate(&pmocker.PMStateDef{EntityType: ms.EntityType, Status: st.Status}).Error; err != nil {
			return err
		}
	}
	return nil
}

// SyncProjectEntities 同步项目实体种子（pm_entities + pm_attrs）
// EPS 模块的 projects 为树层级（仅建 EPS 节点）；业务模块通过 project_code/project_id + entities 建业务实体。
func (s *SeedSyncService) SyncProjectEntities(ctx context.Context, db *gorm.DB, module string, ms *ModuleSeed) error {
	for _, p := range ms.Projects {
		if ms.EntityType == "eps_node" {
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
				if err := s.syncEntity(ctx, db, ms.EntityType, projectID, ent); err != nil {
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

// sortedModuleKeys 模块名排序（EPS 优先，其余按字母序）
func sortedModuleKeys(mods map[string]ModuleSeed) []string {
	keys := make([]string, 0, len(mods))
	for k := range mods {
		keys = append(keys, k)
	}
	// 简单排序：eps 放最前，其余保持插入序（map 遍历无序，但 sync 幂等，顺序影响仅 EPS 先建项目）
	// 用稳定排序保证 EPS 模块优先
	sort.Slice(keys, func(i, j int) bool {
		if keys[i] == "eps" {
			return true
		}
		if keys[j] == "eps" {
			return false
		}
		return keys[i] < keys[j]
	})
	return keys
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
		if v, ok := ent["name"].(string); ok {
			title = v
		}
	}
	if title == "" {
		if v, ok := ent["username"].(string); ok {
			title = v
		}
	}
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
	// nil 值跳过写入：避免 "<nil>" 字符串污染
	if val == nil {
		return nil
	}
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
