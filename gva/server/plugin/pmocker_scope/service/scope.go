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

// CreateScopeItem 创建范围项
func (s *Service) CreateScopeItem(ctx context.Context, projectID uint, title string, attrs map[string]interface{}, creatorID uint) (uint, error) {
	return pmservice.ServiceGroupApp.CreateEntity(ctx, eavtypes.Entity{
		ProjectID:  projectID,
		EntityType: "scope_item",
		Title:      title,
		Status:     "draft",
		CreatedBy:  creatorID,
		Attrs:      attrs,
	})
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
