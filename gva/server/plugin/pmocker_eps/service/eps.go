package eps

import (
	"context"
	"fmt"
	"strings"

	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	pmservice "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
)

type Service struct{}

var EPSService = new(Service)

type TreeNode struct {
	ID       uint       `json:"id"`
	Name     string     `json:"name"`
	Code     string     `json:"code"`
	Type     string     `json:"type"`
	Children []TreeNode `json:"children"`
}

func (s *Service) CreateEPSNode(ctx context.Context, projectID uint, name string, attrs map[string]interface{}, creatorID uint) (uint, error) {
	return pmservice.ServiceGroupApp.CreateEntity(ctx, eavtypes.Entity{
		ProjectID: projectID, EntityType: "eps_node", Title: name, Status: "draft", CreatedBy: creatorID, Attrs: attrs,
	})
}

func (s *Service) GetEPSNode(ctx context.Context, id uint) (*eavtypes.Entity, error) {
	e, err := pmservice.ServiceGroupApp.GetEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	if e.EntityType != "eps_node" {
		return nil, fmt.Errorf("entity %d is not an eps_node", id)
	}
	return e, nil
}

func (s *Service) ListEPSNodes(ctx context.Context, projectID uint, offset, limit int) ([]eavtypes.Entity, int64, error) {
	return pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "eps_node", offset, limit)
}

func (s *Service) UpdateEPSNode(ctx context.Context, e eavtypes.Entity) error {
	if e.EntityType != "eps_node" {
		return fmt.Errorf("not an eps_node")
	}
	return pmservice.ServiceGroupApp.UpdateEntity(ctx, e)
}

func (s *Service) DeleteEPSNode(ctx context.Context, id uint) error {
	return pmservice.ServiceGroupApp.DeleteEntity(ctx, id)
}

func (s *Service) BuildEPSTree(ctx context.Context, projectID uint) ([]TreeNode, error) {
	entities, _, err := pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "eps_node", 0, 10000)
	if err != nil {
		return nil, err
	}
	return buildTreeFromEntities(entities), nil
}

func buildTreeFromEntities(entities []eavtypes.Entity) []TreeNode {
	nodeMap := make(map[string]*TreeNode)

	for _, e := range entities {
		pp := normalizePath(getStrAttr(e.Attrs, "parent_path"))
		name := getStrAttr(e.Attrs, "name")
		if name == "" {
			name = e.Title
		}
		fullPath := joinPath(pp, name)
		nodeMap[fullPath] = &TreeNode{
			ID:       e.ID,
			Name:     name,
			Code:     getStrAttr(e.Attrs, "code"),
			Type:     getStrAttr(e.Attrs, "type"),
			Children: nil, // 下面统一 append，避免再分配
		}
	}

	// 第一遍：用指针把父→子挂到 parentNode.Children（全程指针，避免遍历顺序导致的拷贝丢失）
	// 创建 childrenMap[parentFullPath] = [childNodePointer...]
	childrenMap := make(map[string][]*TreeNode)
	for fullPath, node := range nodeMap {
		parentPath := parentOf(fullPath)
		if _, ok := nodeMap[parentPath]; ok {
			childrenMap[parentPath] = append(childrenMap[parentPath], node)
		}
	}
	// 把挂好的子节点再塞回父节点指针的 Children（转 value 切片）
	for parentFull, children := range childrenMap {
		if pn, ok := nodeMap[parentFull]; ok {
			for _, cn := range children {
				pn.Children = append(pn.Children, *cn)
			}
		}
	}

	// 第二遍：收集根节点（父节点不在 nodeMap 中的），此时根指针的Children已全部就绪
	var roots []TreeNode
	for fullPath, node := range nodeMap {
		parentPath := parentOf(fullPath)
		if _, ok := nodeMap[parentPath]; !ok {
			roots = append(roots, *node)
		}
	}

	return roots
}

// normalizePath 规范化路径，去除首尾多余的斜杠，空路径或/返回空串（根的父节点）
func normalizePath(p string) string {
	p = strings.Trim(p, "/")
	return p
}

// joinPath 拼接 parent_path + name，返回统一格式的全路径（无前导/）
func joinPath(parentNormalized, name string) string {
	if parentNormalized == "" {
		return name
	}
	return parentNormalized + "/" + name
}

// parentOf 返回全路径的父路径（normalize 格式）
func parentOf(fullPath string) string {
	idx := strings.LastIndex(fullPath, "/")
	if idx <= 0 {
		return ""
	}
	return fullPath[:idx]
}

func getStrAttr(attrs map[string]interface{}, key string) string {
	if attrs == nil {
		return ""
	}
	if v, ok := attrs[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (s *Service) AddMember(ctx context.Context, projectID uint, attrs map[string]interface{}, creatorID uint) (uint, error) {
	return pmservice.ServiceGroupApp.CreateEntity(ctx, eavtypes.Entity{
		ProjectID: projectID, EntityType: "eps_member", Title: "EPS成员", Status: "active", CreatedBy: creatorID, Attrs: attrs,
	})
}

func (s *Service) RemoveMember(ctx context.Context, id uint) error {
	return pmservice.ServiceGroupApp.DeleteEntity(ctx, id)
}

func (s *Service) ListMembers(ctx context.Context, projectID uint, offset, limit int) ([]eavtypes.Entity, int64, error) {
	return pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "eps_member", offset, limit)
}

// fullPathOf 根据实体计算规范化全路径（无前导/）
func fullPathOf(e eavtypes.Entity) string {
	pp := normalizePath(getStrAttr(e.Attrs, "parent_path"))
	name := getStrAttr(e.Attrs, "name")
	if name == "" {
		name = e.Title
	}
	return joinPath(pp, name)
}

func (s *Service) MoveNode(ctx context.Context, nodeID uint, newParentPath string) error {
	node, err := s.GetEPSNode(ctx, nodeID)
	if err != nil {
		return err
	}
	oldName := getStrAttr(node.Attrs, "name")
	if oldName == "" {
		oldName = node.Title
	}
	oldParentPathNorm := normalizePath(getStrAttr(node.Attrs, "parent_path"))
	newParentPathNorm := normalizePath(newParentPath)
	oldFull := joinPath(oldParentPathNorm, oldName)
	newFull := joinPath(newParentPathNorm, oldName)

	if err := s.ValidateCycle(ctx, node.ProjectID, oldFull, newFull); err != nil {
		return err
	}

	descendants, err := s.ListDescendants(ctx, node.ProjectID, oldFull)
	if err != nil {
		return err
	}

	if node.Attrs == nil {
		node.Attrs = map[string]interface{}{}
	}
	// 保存回的 parent_path 保留原始格式（前导/），保持数据一致性
	if newParentPathNorm == "" {
		node.Attrs["parent_path"] = "/"
	} else {
		node.Attrs["parent_path"] = "/" + newParentPathNorm
	}
	if err := pmservice.ServiceGroupApp.UpdateEntity(ctx, *node); err != nil {
		return err
	}

	for _, d := range descendants {
		dFull := fullPathOf(d)
		if strings.HasPrefix(dFull, oldFull+"/") {
			rel := strings.TrimPrefix(dFull, oldFull+"/")
			newDFull := newFull + "/" + rel
			newDParent := parentOf(newDFull)
			if d.Attrs == nil {
				d.Attrs = map[string]interface{}{}
			}
			if newDParent == "" {
				d.Attrs["parent_path"] = "/"
			} else {
				d.Attrs["parent_path"] = "/" + newDParent
			}
			if err := pmservice.ServiceGroupApp.UpdateEntity(ctx, d); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) ValidateCycle(ctx context.Context, projectID uint, oldFullPath, newFullPath string) error {
	oldN := normalizePath(oldFullPath)
	newN := normalizePath(newFullPath)
	if newN == oldN || strings.HasPrefix(newN, oldN+"/") {
		return fmt.Errorf("检测到循环引用：不能将节点移动到自身或其子节点下")
	}
	return nil
}

func (s *Service) ListDescendants(ctx context.Context, projectID uint, ancestorFullPath string) ([]eavtypes.Entity, error) {
	entities, _, err := pmservice.ServiceGroupApp.ListEntities(ctx, projectID, "eps_node", 0, 10000)
	if err != nil {
		return nil, err
	}
	ancestorN := normalizePath(ancestorFullPath)
	var result []eavtypes.Entity
	for _, e := range entities {
		fp := fullPathOf(e)
		if ancestorN == "" {
			// 根的后代=所有非根节点
			if parentOf(fp) != "" {
				result = append(result, e)
			}
		} else if strings.HasPrefix(fp, ancestorN+"/") {
			result = append(result, e)
		}
	}
	return result, nil
}
