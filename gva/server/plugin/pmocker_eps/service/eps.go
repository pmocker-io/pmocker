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

// CreateEPSNode 创建 EPS 节点并自动写入 parent_path 属性，保证 BuildEPSTree 能正确构建层级。
// parentID 为 0 表示根节点；非 0 表示作为某 EPS 节点的子节点。
func (s *Service) CreateEPSNode(ctx context.Context, projectID uint, name string, attrs map[string]interface{}, creatorID uint, parentID uint) (uint, error) {
	if attrs == nil {
		attrs = map[string]interface{}{}
	}
	// 确保 name 字段同步到 attrs（buildTreeFromEntities 优先使用 attrs.name，回退到 Title）
	if _, ok := attrs["name"]; !ok {
		attrs["name"] = name
	}
	// 计算 parent_path：
	//   根节点: "/"
	//   子节点: 父节点 parent_path + 父节点 name + "/"
	// buildTreeFromEntities 用 joinPath(parent_path, name) 计算 fullPath，
	// 因此子节点的 parent_path 必须等于父节点的 fullPath。
	parentPath := "/"
	if parentID != 0 {
		parent, err := s.GetEPSNode(ctx, parentID)
		if err != nil {
			return 0, fmt.Errorf("parent eps_node not found: %w", err)
		}
		parentName := ""
		if v, ok := parent.Attrs["name"].(string); ok && v != "" {
			parentName = v
		} else {
			parentName = parent.Title
		}
		parentParentPath := normalizePath(getStrAttr(parent.Attrs, "parent_path"))
		parentPath = joinPath(parentParentPath, parentName)
	}
	attrs["parent_path"] = parentPath
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
	// 保留项目节点 + 用户新建节点；排除组织架构节点（type=group/division，由 GVA 部门管理）
	var projects []eavtypes.Entity
	for _, e := range entities {
		if isOrgNode(e.Attrs) {
			continue
		}
		projects = append(projects, e)
	}
	return buildTreeFromEntities(projects), nil
}

// isOrgNode 判断是否组织架构节点（group=集团/division=部门，由 GVA 部门管理，不在 EPS 树展示）
func isOrgNode(attrs map[string]interface{}) bool {
	if v, ok := attrs["type"].(string); ok {
		if v == "group" || v == "division" {
			return true
		}
	}
	return false
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
			ID:   e.ID,
			Name: name,
			Code: getStrAttr(e.Attrs, "code"),
			Type: getStrAttr(e.Attrs, "type"),
		}
	}

	// 构建 ptrChildren[parentFullPath] = [childPtr...]：全程指针，不做值拷贝
	ptrChildren := make(map[string][]*TreeNode)
	for fullPath, node := range nodeMap {
		parentPath := parentOf(fullPath)
		if _, ok := nodeMap[parentPath]; ok {
			ptrChildren[parentPath] = append(ptrChildren[parentPath], node)
		}
	}

	// 反向指针 -> fullPath，供 clone 递归使用
	ptrToPath := make(map[*TreeNode]string, len(nodeMap))
	for fp, ptr := range nodeMap {
		ptrToPath[ptr] = fp
	}

	// 递归深拷贝：后序顺序天然保证"子节点的Children值切片已完整"后，才被值拷贝到父节点
	// 避免了之前分两步"先把父值拷贝快照，再更新子指针Children造成不同步"的Bug。
	var clone func(fullPath string) TreeNode
	clone = func(fullPath string) TreeNode {
		src := nodeMap[fullPath]
		out := TreeNode{
			ID:   src.ID,
			Name: src.Name,
			Code: src.Code,
			Type: src.Type,
		}
		for _, childPtr := range ptrChildren[fullPath] {
			out.Children = append(out.Children, clone(ptrToPath[childPtr]))
		}
		return out
	}

	var roots []TreeNode
	for fullPath := range nodeMap {
		parentPath := parentOf(fullPath)
		if _, ok := nodeMap[parentPath]; !ok {
			roots = append(roots, clone(fullPath))
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
