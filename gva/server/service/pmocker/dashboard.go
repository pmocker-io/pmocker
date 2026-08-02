package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"gorm.io/gorm"
)

// ProjectDashboard 项目仪表盘聚合数据
type ProjectDashboard struct {
	ProjectID       uint            `json:"projectId"`
	ProjectName     string          `json:"projectName"`
	Priority        int             `json:"priority"`
	Progress        float64         `json:"progress"`
	Health          string          `json:"health"`
	CostSummary     CostSummary     `json:"costSummary"`
	RiskSummary     RiskSummary     `json:"riskSummary"`
	IssueSummary    IssueSummary    `json:"issueSummary"`
	ResourceSummary ResourceSummary `json:"resourceSummary"`
	Milestones      []MilestoneItem `json:"milestones"`
}

type CostSummary struct {
	Budget   float64 `json:"budget"`
	Actual   float64 `json:"actual"`
	Variance float64 `json:"variance"`
	CPI      float64 `json:"cpi"`
}

type RiskSummary struct {
	Total      int            `json:"total"`
	Open       int            `json:"open"`
	High       int            `json:"high"`
	BySeverity map[string]int `json:"bySeverity"`
}

type IssueSummary struct {
	Total    int            `json:"total"`
	Open     int            `json:"open"`
	Closed   int            `json:"closed"`
	ByStatus map[string]int `json:"byStatus"`
}

type ResourceSummary struct {
	MemberCount    int          `json:"memberCount"`
	TotalHours     float64      `json:"totalHours"`
	AvgUtilization float64      `json:"avgUtilization"`
	Members        []MemberUtil `json:"members"`
}

type MemberUtil struct {
	MemberID    uint    `json:"memberId"`
	MemberName  string  `json:"memberName"`
	HourlyRate  float64 `json:"hourlyRate"`
	LoggedHours float64 `json:"loggedHours"`
	Utilization float64 `json:"utilization"`
}

type MilestoneItem struct {
	ID     uint   `json:"id"`
	Title  string `json:"title"`
	Date   string `json:"date"`
	Status string `json:"status"`
}

type DashboardService struct{}

// GetProjectDashboard 聚合项目进度/成本/风险/问题/资源汇总
func (s *DashboardService) GetProjectDashboard(projectID uint) (*ProjectDashboard, error) {
	db := global.GVA_DB

	// 1. 项目基本信息
	var project pmocker.PMEntity
	if err := db.First(&project, projectID).Error; err != nil {
		return nil, err
	}

	dash := &ProjectDashboard{
		ProjectID:   project.ID,
		ProjectName: project.Title,
		Priority:    project.Priority,
	}

	// 2. 进度（任务数简单平均：done/total）
	dash.Progress = s.calcProgressByCount(db, projectID)

	// 3. 成本汇总
	dash.CostSummary = s.calcCostSummary(db, projectID)

	// 4. 风险汇总
	dash.RiskSummary = s.calcRiskSummary(db, projectID)

	// 5. 问题汇总
	dash.IssueSummary = s.calcIssueSummary(db, projectID)

	// 6. 资源汇总
	dash.ResourceSummary = s.calcResourceSummary(db, projectID)

	// 7. 里程碑（任务中标记为 milestone 的）
	dash.Milestones = s.loadMilestones(db, projectID)

	// 8. 健康度：优先使用预设 health 属性，无预设时用计算逻辑
	if h := getAttrString(db, projectID, "health"); h != "" {
		dash.Health = h
	} else {
		dash.Health = s.calcHealthSimple(dash.Progress, dash.CostSummary, dash.RiskSummary)
	}

	return dash, nil
}

// calcProgressByCount 任务数简单平均算法
func (s *DashboardService) calcProgressByCount(db *gorm.DB, projectID uint) float64 {
	var total, done int64
	db.Model(&pmocker.PMEntity{}).
		Where("project_id = ? AND entity_type = ?", projectID, "task").
		Count(&total)
	if total == 0 {
		return 0
	}
	db.Model(&pmocker.PMEntity{}).
		Where("project_id = ? AND entity_type = ? AND status = ?", projectID, "task", "done").
		Count(&done)
	return float64(done) / float64(total) * 100
}

// calcCostSummary 成本汇总：预算来自 cost_item 实体的 planned_value attr，实际来自 pm_cost_actuals
func (s *DashboardService) calcCostSummary(db *gorm.DB, projectID uint) CostSummary {
	var cs CostSummary
	// 预算：cost_item 实体的 planned_value attr 求和
	var costItems []pmocker.PMEntity
	db.Where("project_id = ? AND entity_type = ?", projectID, "cost_item").Find(&costItems)
	for _, ci := range costItems {
		cs.Budget += getAttrDecimal(db, ci.ID, "planned_value")
	}
	// 实际：pm_cost_actuals 求和
	var actualSum float64
	db.Model(&pmocker.PMCostActual{}).
		Where("project_id = ? AND status = ?", projectID, "confirmed").
		Select("COALESCE(SUM(amount),0)").Scan(&actualSum)
	cs.Actual = actualSum
	cs.Variance = cs.Actual - cs.Budget
	if cs.Actual > 0 {
		cs.CPI = cs.Budget / cs.Actual
	}
	return cs
}

// calcRiskSummary 风险汇总
func (s *DashboardService) calcRiskSummary(db *gorm.DB, projectID uint) RiskSummary {
	var risks []pmocker.PMEntity
	db.Where("project_id = ? AND entity_type = ?", projectID, "risk").Find(&risks)
	rs := RiskSummary{BySeverity: map[string]int{}}
	for _, r := range risks {
		rs.Total++
		severity := getAttrString(db, r.ID, "risk_level")
		if severity == "" {
			severity = "medium"
		}
		rs.BySeverity[severity]++
		if r.Status != "closed" {
			rs.Open++
			if severity == "high" {
				rs.High++
			}
		}
	}
	return rs
}

// calcIssueSummary 问题汇总
func (s *DashboardService) calcIssueSummary(db *gorm.DB, projectID uint) IssueSummary {
	var issues []pmocker.PMEntity
	db.Where("project_id = ? AND entity_type = ?", projectID, "issue").Find(&issues)
	is := IssueSummary{ByStatus: map[string]int{}}
	for _, i := range issues {
		is.Total++
		is.ByStatus[i.Status]++
		if i.Status == "closed" || i.Status == "resolved" {
			is.Closed++
		} else {
			is.Open++
		}
	}
	return is
}

// calcResourceSummary 资源汇总
func (s *DashboardService) calcResourceSummary(db *gorm.DB, projectID uint) ResourceSummary {
	var members []pmocker.PMEntity
	db.Where("project_id = ? AND entity_type = ?", projectID, "team_member").Find(&members)
	rs := ResourceSummary{Members: []MemberUtil{}}
	if len(members) == 0 {
		return rs
	}
	rs.MemberCount = len(members)
	totalUtil := 0.0
	for _, m := range members {
		rate := getAttrDecimal(db, m.ID, "hourly_rate")
		// 工时来自 pm_time_entries（approved）
		var hours float64
		db.Model(&pmocker.PMTimeEntry{}).
			Where("member_id = ? AND status = ?", m.ID, "approved").
			Select("COALESCE(SUM(hours),0)").Scan(&hours)
		util := 0.0
		if hours > 0 {
			util = hours / 160.0 * 100 // 月计划工时160h
		}
		rs.TotalHours += hours
		totalUtil += util
		rs.Members = append(rs.Members, MemberUtil{
			MemberID:    m.ID,
			MemberName:  m.Title,
			HourlyRate:  rate,
			LoggedHours: hours,
			Utilization: util,
		})
	}
	rs.AvgUtilization = totalUtil / float64(len(members))
	return rs
}

// loadMilestones 里程碑（标记 milestone=1 的任务）
func (s *DashboardService) loadMilestones(db *gorm.DB, projectID uint) []MilestoneItem {
	var tasks []pmocker.PMEntity
	db.Where("project_id = ? AND entity_type = ?", projectID, "task").Find(&tasks)
	var ms []MilestoneItem
	for _, t := range tasks {
		if getAttrInt(db, t.ID, "is_milestone") == 1 {
			ms = append(ms, MilestoneItem{
				ID:     t.ID,
				Title:  t.Title,
				Date:   getAttrString(db, t.ID, "end_date"),
				Status: t.Status,
			})
		}
	}
	return ms
}

// calcHealthSimple 基于进度/成本/风险的简单健康度
func (s *DashboardService) calcHealthSimple(progress float64, cs CostSummary, rs RiskSummary) string {
	// 成本偏差率
	costVarPct := 0.0
	if cs.Budget > 0 {
		costVarPct = (cs.Actual - cs.Budget) / cs.Budget * 100
	}
	if costVarPct > 20 || rs.High >= 3 {
		return "red"
	}
	if costVarPct > 10 || rs.High >= 1 {
		return "yellow"
	}
	return "green"
}

// ===== EAV attr 读取 helper（包内私有，后续 Task 共用）=====

func getAttrDecimal(db *gorm.DB, entityID uint, key string) float64 {
	var attr pmocker.PMAttr
	db.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr)
	if attr.ValDecimal != nil {
		return *attr.ValDecimal
	}
	return 0
}

func getAttrInt(db *gorm.DB, entityID uint, key string) int64 {
	var attr pmocker.PMAttr
	db.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr)
	if attr.ValInt != nil {
		return *attr.ValInt
	}
	return 0
}

func getAttrString(db *gorm.DB, entityID uint, key string) string {
	var attr pmocker.PMAttr
	db.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr)
	if attr.ValString != nil {
		return *attr.ValString
	}
	return ""
}

func getAttrRef(db *gorm.DB, entityID uint, key string) uint {
	var attr pmocker.PMAttr
	db.Where("entity_id = ? AND field_key = ?", entityID, key).First(&attr)
	if attr.ValRef != nil {
		return *attr.ValRef
	}
	return 0
}
