package pmocker

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"gorm.io/gorm"
)

// MyTaskItem 个人任务项（统一 4 类来源）
type MyTaskItem struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	SourceType  string `json:"sourceType"` // project_task/issue_task/change_task/deliverable_task
	ProjectID   uint   `json:"projectId"`
	ProjectName string `json:"projectName"`
	Status      string `json:"status"`
	Priority    int    `json:"priority"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Progress    int    `json:"progress"`
	OwnerID     uint   `json:"ownerId"`
	Overdue     bool   `json:"overdue"`
}

// TaskStats 任务统计
type TaskStats struct {
	Total        int     `json:"total"`
	Done         int     `json:"done"`
	DoneRate     float64 `json:"doneRate"`
	OverdueCount int     `json:"overdueCount"`
	TodoCount    int     `json:"todoCount"`
	DoingCount   int     `json:"doingCount"`
}

type TaskCenterService struct{}

// GetMyTasks 聚合项目任务(owner_id)+问题任务(assignee)+变更任务+交付物任务(reviewer)
func (s *TaskCenterService) GetMyTasks(userID uint) ([]MyTaskItem, error) {
	db := global.GVA_DB
	var tasks []MyTaskItem

	// 1. 项目任务：pm_entities.owner_id = userID 且 entity_type=task
	tasks = append(tasks, s.loadProjectTasks(db, userID)...)
	// 2. 问题任务：EAV attr assignee = userID
	tasks = append(tasks, s.loadAttrAssignedTasks(db, userID, "issue", "issue_task", "assignee")...)
	// 3. 变更任务：EAV attr assignee = userID
	tasks = append(tasks, s.loadAttrAssignedTasks(db, userID, "change", "change_task", "assignee")...)
	// 4. 交付物任务：EAV attr reviewer = userID
	tasks = append(tasks, s.loadAttrAssignedTasks(db, userID, "deliverable", "deliverable_task", "reviewer")...)

	s.enrichOverdue(tasks)
	return tasks, nil
}

// GetMyTasksByStatus 按状态分组：todo/doing/done/overdue
func (s *TaskCenterService) GetMyTasksByStatus(userID uint, status string) ([]MyTaskItem, error) {
	all, err := s.GetMyTasks(userID)
	if err != nil {
		return nil, err
	}
	var filtered []MyTaskItem
	for _, t := range all {
		group := s.mapStatusGroup(t)
		if group == status {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

// GetMyFocusedTasks P0/P1 高优先级任务，按可见性规则
// PMO_ADMIN: 可见所有 P0/P1 任务
// DEPT_LEADER: 可见本部门及子级下 P0/P1 任务
// CCB_MEMBER: 可见所有 P0/P1 变更相关任务
// 其他: 仅见自己负责的 P0/P1 任务
func (s *TaskCenterService) GetMyFocusedTasks(userID uint) ([]MyTaskItem, error) {
	db := global.GVA_DB
	scope := getUserVisibilityScope(db, userID)

	var tasks []MyTaskItem

	if scope.IsPMOAdmin {
		// 可见所有 P0/P1 任务（4 类来源）
		tasks = append(tasks, s.loadProjectTasksByPriority(db, nil, 0, 1)...)
		tasks = append(tasks, s.loadAttrAssignedTasksByPriority(db, nil, "issue", "issue_task", "assignee", 0, 1)...)
		tasks = append(tasks, s.loadAttrAssignedTasksByPriority(db, nil, "change", "change_task", "assignee", 0, 1)...)
		tasks = append(tasks, s.loadAttrAssignedTasksByPriority(db, nil, "deliverable", "deliverable_task", "reviewer", 0, 1)...)
	} else {
		// 自己负责的 P0/P1 任务（所有角色默认可见）
		tasks = append(tasks, s.loadProjectTasksByPriority(db, &userID, 0, 1)...)
		tasks = append(tasks, s.loadAttrAssignedTasksByPriority(db, &userID, "issue", "issue_task", "assignee", 0, 1)...)
		tasks = append(tasks, s.loadAttrAssignedTasksByPriority(db, &userID, "change", "change_task", "assignee", 0, 1)...)
		tasks = append(tasks, s.loadAttrAssignedTasksByPriority(db, &userID, "deliverable", "deliverable_task", "reviewer", 0, 1)...)

		// CCB_MEMBER: 可见所有 P0/P1 变更任务（在自负责之外补充）
		if scope.IsCCBMember {
			tasks = append(tasks, s.loadAttrAssignedTasksByPriority(db, nil, "change", "change_task", "assignee", 0, 1)...)
		}

		// DEPT_LEADER: 可见本部门及子级下 P0/P1 任务
		if scope.IsDeptLeader && scope.DeptID > 0 {
			tasks = append(tasks, s.loadDeptTasksByPriority(db, scope, 0, 1)...)
		}
	}

	s.enrichOverdue(tasks)
	return tasks, nil
}

// GetTaskStats 统计任务总数/完成率/逾期数
func (s *TaskCenterService) GetTaskStats(userID uint) (*TaskStats, error) {
	all, err := s.GetMyTasks(userID)
	if err != nil {
		return nil, err
	}
	st := &TaskStats{Total: len(all)}
	today := time.Now().Format("2006-01-02")
	for _, t := range all {
		group := s.mapStatusGroup(t)
		switch group {
		case "todo":
			st.TodoCount++
		case "doing":
			st.DoingCount++
		case "done":
			st.Done++
		}
		if t.EndDate != "" && t.EndDate < today && group != "done" {
			st.OverdueCount++
		}
	}
	if st.Total > 0 {
		st.DoneRate = float64(st.Done) / float64(st.Total) * 100
	}
	return st, nil
}

// ===== 内部加载方法 =====

// loadProjectTasks 加载 pm_entities.owner_id = userID 的任务
func (s *TaskCenterService) loadProjectTasks(db *gorm.DB, userID uint) []MyTaskItem {
	var entities []pmocker.PMEntity
	db.Where("entity_type = ? AND owner_id = ?", "task", userID).Find(&entities)
	return s.toMyTasks(db, entities, "project_task")
}

// loadProjectTasksByPriority 按优先级范围加载项目任务，userID=nil 表示所有
func (s *TaskCenterService) loadProjectTasksByPriority(db *gorm.DB, userID *uint, minP, maxP int) []MyTaskItem {
	q := db.Model(&pmocker.PMEntity{}).Where("entity_type = ? AND priority BETWEEN ? AND ?", "task", minP, maxP)
	if userID != nil {
		q = q.Where("owner_id = ?", *userID)
	}
	var entities []pmocker.PMEntity
	q.Find(&entities)
	return s.toMyTasks(db, entities, "project_task")
}

// loadAttrAssignedTasks 加载 EAV attr 指定字段 = userID 的实体
func (s *TaskCenterService) loadAttrAssignedTasks(db *gorm.DB, userID uint, entityType, sourceType, attrKey string) []MyTaskItem {
	var entities []pmocker.PMEntity
	db.Joins("JOIN pm_attrs ON pm_attrs.entity_id = pm_entities.id").
		Where("pm_entities.entity_type = ? AND pm_attrs.field_key = ? AND pm_attrs.val_int = ?",
			entityType, attrKey, userID).
		Find(&entities)
	return s.toMyTasks(db, entities, sourceType)
}

// loadAttrAssignedTasksByPriority 按优先级范围加载
func (s *TaskCenterService) loadAttrAssignedTasksByPriority(db *gorm.DB, userID *uint, entityType, sourceType, attrKey string, minP, maxP int) []MyTaskItem {
	q := db.Model(&pmocker.PMEntity{}).
		Joins("JOIN pm_attrs ON pm_attrs.entity_id = pm_entities.id").
		Where("pm_entities.entity_type = ? AND pm_attrs.field_key = ? AND pm_entities.priority BETWEEN ? AND ?",
			entityType, attrKey, minP, maxP)
	if userID != nil {
		q = q.Where("pm_attrs.val_int = ?", *userID)
	}
	var entities []pmocker.PMEntity
	q.Find(&entities)
	return s.toMyTasks(db, entities, sourceType)
}

// loadDeptTasksByPriority 部门负责人可见本部门及子级下 P0/P1 任务
func (s *TaskCenterService) loadDeptTasksByPriority(db *gorm.DB, scope VisibilityScope, minP, maxP int) []MyTaskItem {
	// 查本部门及子部门下的用户
	var userIDs []uint
	db.Table("sys_users").
		Where("dept_id = ? OR dept_id IN (SELECT id FROM sys_departments WHERE ancestors LIKE ?)",
			scope.DeptID, scope.DeptAncestors+"%").
		Pluck("id", &userIDs)
	if len(userIDs) == 0 {
		return nil
	}
	var entities []pmocker.PMEntity
	db.Where("entity_type = ? AND priority BETWEEN ? AND ? AND owner_id IN ?",
		"task", minP, maxP, userIDs).Find(&entities)
	return s.toMyTasks(db, entities, "project_task")
}

// toMyTasks 实体转 MyTaskItem（含项目名/进度/日期）
func (s *TaskCenterService) toMyTasks(db *gorm.DB, entities []pmocker.PMEntity, sourceType string) []MyTaskItem {
	var items []MyTaskItem
	for _, e := range entities {
		item := MyTaskItem{
			ID:         e.ID,
			Title:      e.Title,
			SourceType: sourceType,
			ProjectID:  e.ProjectID,
			Status:     e.Status,
			Priority:   e.Priority,
			StartDate:  getAttrString(db, e.ID, "start_date"),
			EndDate:    getAttrString(db, e.ID, "end_date"),
			Progress:   int(getAttrInt(db, e.ID, "progress")),
		}
		if e.OwnerID != nil {
			item.OwnerID = *e.OwnerID
		}
		// 项目名
		var proj pmocker.PMEntity
		db.First(&proj, e.ProjectID)
		item.ProjectName = proj.Title
		items = append(items, item)
	}
	return items
}

// enrichOverdue 填充逾期标记
func (s *TaskCenterService) enrichOverdue(tasks []MyTaskItem) {
	today := time.Now().Format("2006-01-02")
	for i := range tasks {
		if tasks[i].EndDate != "" && tasks[i].EndDate < today && s.mapStatusGroup(tasks[i]) != "done" {
			tasks[i].Overdue = true
		}
	}
}

// mapStatusGroup 状态映射到分组：todo/doing/done/overdue
func (s *TaskCenterService) mapStatusGroup(t MyTaskItem) string {
	switch t.Status {
	case "done", "closed", "completed", "approved", "published", "archived", "resolved":
		return "done"
	case "doing", "in_progress", "processing", "reviewing":
		return "doing"
	default: // todo, open, draft, pending, new
		return "todo"
	}
}
