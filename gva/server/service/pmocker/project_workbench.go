package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"gorm.io/gorm"
)

// MyProjects 我创建的/我负责的/我参与的
type MyProjects struct {
	Created  []ProjectCard `json:"created"`  // 我创建的
	Lead     []ProjectCard `json:"lead"`     // 我负责的（leader_id=我）
	Involved []ProjectCard `json:"involved"` // 我参与的（team_member.user_id=我）
}

type ProjectWorkbenchService struct{}

// GetMyProjects 分三类返回
func (s *ProjectWorkbenchService) GetMyProjects(userID uint) (*MyProjects, error) {
	db := global.GVA_DB
	result := &MyProjects{
		Created: []ProjectCard{}, Lead: []ProjectCard{}, Involved: []ProjectCard{},
	}

	// 我创建的：pm_entities.created_by = userID 且 entity_type=eps_node（仅项目，排除 EPS 组织节点）
	var created []pmocker.PMEntity
	db.Where("entity_type = ? AND created_by = ? AND id IN (SELECT entity_id FROM pm_attrs WHERE field_key = ?)", "eps_node", userID, "progress_algo").Find(&created)
	for _, p := range created {
		if card, err := (&PMODashboardService{}).GetProjectCard(p.ID); err == nil {
			result.Created = append(result.Created, *card)
		}
	}

	// 我负责的：owner_id = userID 且 entity_type=eps_node（仅项目）
	var lead []pmocker.PMEntity
	db.Where("entity_type = ? AND owner_id = ? AND id IN (SELECT entity_id FROM pm_attrs WHERE field_key = ?)", "eps_node", userID, "progress_algo").Find(&lead)
	for _, p := range lead {
		if card, err := (&PMODashboardService{}).GetProjectCard(p.ID); err == nil {
			result.Lead = append(result.Lead, *card)
		}
	}

	// 我参与的：team_member.owner_id = userID（team_member 关联 sys_user）→ 取 project_id
	var involvedProjectIDs []uint
	db.Model(&pmocker.PMEntity{}).
		Where("entity_type = ? AND owner_id = ?", "team_member", userID).
		Distinct("project_id").Pluck("project_id", &involvedProjectIDs)
	for _, pid := range involvedProjectIDs {
		if card, err := (&PMODashboardService{}).GetProjectCard(pid); err == nil {
			result.Involved = append(result.Involved, *card)
		}
	}

	return result, nil
}

// GetMyProjectsByStatus 按状态分组：initiating/active/archived/paused
func (s *ProjectWorkbenchService) GetMyProjectsByStatus(userID uint, status string) ([]ProjectCard, error) {
	all, err := s.GetMyProjects(userID)
	if err != nil {
		return nil, err
	}
	// 合并去重
	seen := map[uint]bool{}
	var merged []ProjectCard
	for _, list := range [][]ProjectCard{all.Created, all.Lead, all.Involved} {
		for _, c := range list {
			if !seen[c.ProjectID] {
				seen[c.ProjectID] = true
				merged = append(merged, c)
			}
		}
	}
	// status 映射：initiating=planning/draft, active=active/doing, archived=archived, paused=paused/on_hold
	var filtered []ProjectCard
	for _, c := range merged {
		if s.mapProjectStatusGroup(c.Status) == status {
			filtered = append(filtered, c)
		}
	}
	return filtered, nil
}

// GetMyFocusedProjects P0/P1 高优先级项目，按可见性规则
// PMO_ADMIN: 可见所有 P0/P1 项目
// DEPT_LEADER: 可见本部门及子级下 P0/P1 项目
// 其他: 仅见自己参与的 P0/P1 项目
func (s *ProjectWorkbenchService) GetMyFocusedProjects(userID uint) ([]ProjectCard, error) {
	db := global.GVA_DB
	scope := getUserVisibilityScope(db, userID)

	var projects []pmocker.PMEntity
	q := db.Where("entity_type = ? AND priority BETWEEN ? AND ? AND id IN (SELECT entity_id FROM pm_attrs WHERE field_key = ?)", "eps_node", 0, 1, "progress_algo")
	if !scope.IsPMOAdmin {
		if scope.IsDeptLeader && scope.DeptID > 0 {
			// 部门负责人：本部门及子级下的项目
			deptIDs := s.getDeptAndChildren(db, scope.DeptID, scope.DeptAncestors)
			q = q.Where("id IN (SELECT entity_id FROM pm_attrs WHERE field_key = ? AND val_int IN ?)", "dept_id", deptIDs)
		} else {
			// 其他角色：自己参与的项目（创建/负责/团队成员）
			myProjectIDs := s.getMyInvolvedProjectIDs(db, userID)
			if len(myProjectIDs) == 0 {
				return []ProjectCard{}, nil
			}
			q = q.Where("id IN ?", myProjectIDs)
		}
	}
	q.Find(&projects)

	var cards []ProjectCard
	for _, p := range projects {
		if card, err := (&PMODashboardService{}).GetProjectCard(p.ID); err == nil {
			cards = append(cards, *card)
		}
	}
	return cards, nil
}

// ===== 内部 helper =====

// mapProjectStatusGroup 项目状态映射
func (s *ProjectWorkbenchService) mapProjectStatusGroup(status string) string {
	switch status {
	case "planning", "draft", "initiating":
		return "initiating"
	case "archived":
		return "archived"
	case "paused", "on_hold":
		return "paused"
	default: // active, doing, in_progress
		return "active"
	}
}

// getDeptAndChildren 获取部门及子部门ID
func (s *ProjectWorkbenchService) getDeptAndChildren(db *gorm.DB, deptID uint, ancestors string) []uint {
	var ids []uint
	ids = append(ids, deptID)
	var childIDs []uint
	db.Table("sys_departments").
		Where("ancestors LIKE ?", ancestors+"%").
		Pluck("id", &childIDs)
	ids = append(ids, childIDs...)
	return ids
}

// getMyInvolvedProjectIDs 获取我参与的项目ID（创建/负责/团队成员）
func (s *ProjectWorkbenchService) getMyInvolvedProjectIDs(db *gorm.DB, userID uint) []uint {
	var ids []uint
	// 创建的 + 负责的（仅项目，排除 EPS 组织节点）
	db.Model(&pmocker.PMEntity{}).
		Where("entity_type = ? AND (created_by = ? OR owner_id = ?) AND id IN (SELECT entity_id FROM pm_attrs WHERE field_key = ?)", "eps_node", userID, userID, "progress_algo").
		Pluck("id", &ids)
	// 团队成员参与的
	var memberProjectIDs []uint
	db.Model(&pmocker.PMEntity{}).
		Where("entity_type = ? AND owner_id = ?", "team_member", userID).
		Distinct("project_id").Pluck("project_id", &memberProjectIDs)
	ids = append(ids, memberProjectIDs...)
	return ids
}
