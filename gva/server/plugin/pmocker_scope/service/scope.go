package scope

import (
	"context"
	"fmt"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	pmservice "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
)

type Service struct{}

var ScopeService = new(Service)

// CreateScopeItem 创建范围项并同时插入 WBS 节点（保证 GetWBS 能查到）。
// parentID 为 0 表示根节点；非 0 表示作为某 WBS 节点的子节点。
func (s *Service) CreateScopeItem(ctx context.Context, projectID uint, title string, attrs map[string]interface{}, creatorID uint, parentID uint) (uint, error) {
	entityID, err := pmservice.ServiceGroupApp.CreateEntity(ctx, eavtypes.Entity{
		ProjectID:  projectID,
		EntityType: "scope_item",
		Title:      title,
		Status:     "draft",
		CreatedBy:  creatorID,
		Attrs:      attrs,
	})
	if err != nil {
		return 0, err
	}
	// 创建 WBS 节点并关联到新建的 scope_item
	if _, err := s.ensureWBSNode(ctx, projectID, parentID, entityID); err != nil {
		// WBS 节点创建失败不影响 entity 已创建的事实，前端可重试或通过 updateScopeItem 补建
		return entityID, fmt.Errorf("scope_item created(id=%d) but wbs node failed: %w", entityID, err)
	}
	return entityID, nil
}

// ensureWBSNode 创建 WBS 节点：parentID==0 为根节点，否则为子节点。
func (s *Service) ensureWBSNode(ctx context.Context, projectID uint, parentID uint, entityID uint) (uint, error) {
	if parentID == 0 {
		// 根节点：path = 当前项目根节点数量+1, level = 1
		var rootCount int64
		if err := global.GVA_DB.WithContext(ctx).Model(&pmocker.PMWBSNode{}).
			Where("project_id = ? AND parent_id IS NULL", projectID).Count(&rootCount).Error; err != nil {
			return 0, err
		}
		node := pmocker.PMWBSNode{
			ProjectID: projectID,
			Path:      fmt.Sprintf("%d", rootCount+1),
			Level:     1,
			EntityID:  entityID,
		}
		if err := global.GVA_DB.WithContext(ctx).Create(&node).Error; err != nil {
			return 0, err
		}
		return node.ID, nil
	}
	// 子节点：依赖父节点 path/level 推导
	return s.AddWBSChild(ctx, parentID, entityID)
}

// ListScopeItems 列出项目下的范围项
func (s *Service) ListScopeItems(ctx context.Context, projectID uint, offset, limit int) ([]eavtypes.Entity, int64, error) {
	return pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "scope_item", offset, limit)
}

// BuildWBS 构建项目 WBS 树
func (s *Service) BuildWBS(ctx context.Context, projectID uint, rootTitle string) (uint, error) {
	root := pmocker.PMWBSNode{
		ProjectID: projectID,
		Path:      "1",
		Level:     1,
	}
	if err := global.GVA_DB.WithContext(ctx).Create(&root).Error; err != nil {
		return 0, err
	}
	return root.ID, nil
}

// AddWBSChild 添加 WBS 子节点
func (s *Service) AddWBSChild(ctx context.Context, parentID uint, scopeItemID uint) (uint, error) {
	var parent pmocker.PMWBSNode
	if err := global.GVA_DB.WithContext(ctx).First(&parent, parentID).Error; err != nil {
		return 0, fmt.Errorf("parent node not found: %w", err)
	}
	// 查询兄弟节点数决定编号
	var count int64
	global.GVA_DB.WithContext(ctx).Model(&pmocker.PMWBSNode{}).
		Where("parent_id = ?", parentID).Count(&count)
	newPath := fmt.Sprintf("%s.%d", parent.Path, count+1)
	child := pmocker.PMWBSNode{
		ProjectID: parent.ProjectID,
		ParentID:  &parentID,
		Path:      newPath,
		Level:     parent.Level + 1,
		EntityID:  scopeItemID,
	}
	if err := global.GVA_DB.WithContext(ctx).Create(&child).Error; err != nil {
		return 0, err
	}
	return child.ID, nil
}

// GetWBS 获取项目 WBS 树
func (s *Service) GetWBS(ctx context.Context, projectID uint) ([]WBSNode, error) {
	var nodes []pmocker.PMWBSNode
	if err := global.GVA_DB.WithContext(ctx).Where("project_id = ?", projectID).Order("path").Find(&nodes).Error; err != nil {
		return nil, err
	}
	result := make([]WBSNode, len(nodes))
	for i, n := range nodes {
		result[i] = WBSNode{
			ID:          n.ID,
			ParentID:    n.ParentID,
			Path:        n.Path,
			Level:       n.Level,
			ScopeItemID: n.EntityID,
		}
	}
	return result, nil
}

// SetBaseline 设置范围基线
func (s *Service) SetBaseline(ctx context.Context, projectID uint, snapshotJSON string) (uint, error) {
	bl := pmocker.PMBaseline{
		ProjectID:    projectID,
		Type:         "scope",
		SnapshotJSON: snapshotJSON,
	}
	if err := global.GVA_DB.WithContext(ctx).Create(&bl).Error; err != nil {
		return 0, err
	}
	return bl.ID, nil
}

// WBSNode WBS 节点 DTO
type WBSNode struct {
	ID          uint   `json:"id"`
	ParentID    *uint  `json:"parentId"`
	Path        string `json:"path"`
	Level       int    `json:"level"`
	ScopeItemID uint   `json:"scopeItemId"`
}

// PathLevel 从物化路径解析层级
func PathLevel(path string) int {
	return strings.Count(path, ".") + 1
}
