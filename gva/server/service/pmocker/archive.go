package pmocker

import (
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"gorm.io/gorm"
)

// CloseReport 结项报告
type CloseReport struct {
	ProjectID    uint          `json:"projectId"`
	ProjectName  string        `json:"projectName"`
	StartDate    string        `json:"startDate"`
	EndDate      string        `json:"endDate"`
	ArchivedAt   string        `json:"archivedAt"`
	TaskStat     CategoryStat  `json:"taskStat"`
	IssueStat    CategoryStat  `json:"issueStat"`
	RiskStat     CategoryStat  `json:"riskStat"`
	ReqStat      CategoryStat  `json:"reqStat"`
	ChangeStat   CategoryStat  `json:"changeStat"`
	ResourceStat ResourceStat  `json:"resourceStat"`
	CostStat     CostStat      `json:"costStat"`
}

type CategoryStat struct {
	Total   int `json:"total"`
	Done    int `json:"done"`
	Open    int `json:"open"`
	Closed  int `json:"closed"`
	Overdue int `json:"overdue"`
}

type ResourceStat struct {
	MemberCount int     `json:"memberCount"`
	TotalHours  float64 `json:"totalHours"`
	TotalCost   float64 `json:"totalCost"`
}

type CostStat struct {
	Budget   float64 `json:"budget"`
	Actual   float64 `json:"actual"`
	Variance float64 `json:"variance"`
}

type ArchiveService struct{}

// ArchiveProject 归档状态机
// 1. 验证所有任务已完成、所有问题已关闭、所有交付物已归档
// 2. 生成结项报告快照
// 3. 设置项目 status=archived，所有关联实体标记 archived
func (s *ArchiveService) ArchiveProject(projectID uint, archivedBy uint) error {
	db := global.GVA_DB

	// 1. 验证前置条件
	if err := s.validateArchivable(db, projectID); err != nil {
		return err
	}

	// 2. 生成结项报告快照
	rs := &ReportService{}
	if err := rs.GenerateReportSnapshot(projectID, "close", "close", archivedBy); err != nil {
		return fmt.Errorf("生成结项报告快照失败: %w", err)
	}

	// 3. 设置项目状态为 archived
	now := time.Now().Format("2006-01-02 15:04:05")
	if err := db.Model(&pmocker.PMEntity{}).Where("id = ?", projectID).
		Updates(map[string]interface{}{"status": "archived", "updated_at": now}).Error; err != nil {
		return err
	}

	// 4. 所有关联实体标记 archived（状态标记，不使用 gorm 软删除以保持报告可查）
	entityTypes := []string{"task", "issue", "risk", "requirement", "change", "deliverable",
		"team_member", "cost_item", "scope_item"}
	for _, et := range entityTypes {
		db.Model(&pmocker.PMEntity{}).
			Where("project_id = ? AND entity_type = ? AND status != ?", projectID, et, "archived").
			Update("status", "archived")
	}

	return nil
}

// validateArchivable 验证项目可归档
func (s *ArchiveService) validateArchivable(db *gorm.DB, projectID uint) error {
	// 任务全部完成
	var undoneTasks int64
	db.Model(&pmocker.PMEntity{}).
		Where("project_id = ? AND entity_type = ? AND status NOT IN ?", projectID, "task", []string{"done", "archived"}).
		Count(&undoneTasks)
	if undoneTasks > 0 {
		return fmt.Errorf("存在 %d 个未完成任务，无法归档", undoneTasks)
	}
	// 问题全部关闭
	var openIssues int64
	db.Model(&pmocker.PMEntity{}).
		Where("project_id = ? AND entity_type = ? AND status NOT IN ?", projectID, "issue", []string{"closed", "resolved", "archived"}).
		Count(&openIssues)
	if openIssues > 0 {
		return fmt.Errorf("存在 %d 个未关闭问题，无法归档", openIssues)
	}
	// 交付物全部归档/发布
	var unarchivedDeliverables int64
	db.Model(&pmocker.PMEntity{}).
		Where("project_id = ? AND entity_type = ? AND status NOT IN ?", projectID, "deliverable", []string{"archived", "published"}).
		Count(&unarchivedDeliverables)
	if unarchivedDeliverables > 0 {
		return fmt.Errorf("存在 %d 个未归档交付物，无法归档", unarchivedDeliverables)
	}
	return nil
}

// GetCloseReport 生成结项报告数据
func (s *ArchiveService) GetCloseReport(projectID uint) (*CloseReport, error) {
	db := global.GVA_DB
	var project pmocker.PMEntity
	if err := db.First(&project, projectID).Error; err != nil {
		return nil, err
	}

	report := &CloseReport{
		ProjectID:   project.ID,
		ProjectName: project.Title,
		StartDate:   getAttrString(db, projectID, "start_date"),
		EndDate:     getAttrString(db, projectID, "end_date"),
	}

	today := time.Now().Format("2006-01-02")
	report.TaskStat = s.categoryStat(db, projectID, "task", []string{"done"}, today)
	report.IssueStat = s.categoryStat(db, projectID, "issue", []string{"closed", "resolved"}, today)
	report.RiskStat = s.categoryStat(db, projectID, "risk", []string{"closed"}, today)
	report.ReqStat = s.categoryStat(db, projectID, "requirement", []string{"approved", "implemented"}, today)
	report.ChangeStat = s.categoryStat(db, projectID, "change", []string{"approved", "closed"}, today)

	// 资源统计
	var memberCount int64
	db.Model(&pmocker.PMEntity{}).Where("project_id = ? AND entity_type = ?", projectID, "team_member").Count(&memberCount)
	var totalHours float64
	db.Model(&pmocker.PMTimeEntry{}).Where("project_id = ? AND status = ?", projectID, "approved").
		Select("COALESCE(SUM(hours),0)").Scan(&totalHours)
	var totalCost float64
	db.Model(&pmocker.PMCostActual{}).Where("project_id = ? AND status = ?", projectID, "confirmed").
		Select("COALESCE(SUM(amount),0)").Scan(&totalCost)
	report.ResourceStat = ResourceStat{
		MemberCount: int(memberCount),
		TotalHours:  totalHours,
		TotalCost:   totalCost,
	}

	// 成本统计
	ds := &DashboardService{}
	cs := ds.calcCostSummary(db, projectID)
	report.CostStat = CostStat{Budget: cs.Budget, Actual: cs.Actual, Variance: cs.Variance}

	if project.Status == "archived" {
		report.ArchivedAt = project.UpdatedAt.Format("2006-01-02 15:04:05")
	}
	return report, nil
}

// categoryStat 按实体类型统计
func (s *ArchiveService) categoryStat(db *gorm.DB, projectID uint, entityType string, doneStatuses []string, today string) CategoryStat {
	var entities []pmocker.PMEntity
	db.Where("project_id = ? AND entity_type = ?", projectID, entityType).Find(&entities)
	st := CategoryStat{Total: len(entities)}
	for _, e := range entities {
		if contains(doneStatuses, e.Status) || e.Status == "archived" {
			st.Done++
		} else {
			st.Open++
		}
		if e.Status == "closed" || e.Status == "archived" {
			st.Closed++
		}
		// 逾期：end_date < today 且未完成
		endDate := getAttrString(db, e.ID, "end_date")
		if endDate != "" && endDate < today && !contains(doneStatuses, e.Status) {
			st.Overdue++
		}
	}
	return st
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
