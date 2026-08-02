package pmocker

import (
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"gorm.io/gorm"
)

// PMODashboard PMO 看板聚合数据
type PMODashboard struct {
	TotalProjects int            `json:"totalProjects"`
	HealthDist    map[string]int `json:"healthDist"` // green/yellow/red 计数
	ProjectCards  []ProjectCard  `json:"projectCards"`
	LoadSummary   LoadSummary    `json:"loadSummary"`
}

// ProjectCard 项目卡片数据（PMO看板 + 项目工作台共用）
type ProjectCard struct {
	ProjectID    uint    `json:"projectId"`
	ProjectName  string  `json:"projectName"`
	Priority     int     `json:"priority"`
	Status       string  `json:"status"`
	Health       string  `json:"health"`
	Progress     float64 `json:"progress"`
	Budget       float64 `json:"budget"`
	Actual       float64 `json:"actual"`
	CostVariance float64 `json:"costVariance"`
	RiskCount    int     `json:"riskCount"`
	LeaderName   string  `json:"leaderName"`
	DeptID       uint    `json:"deptId"`
}

type LoadSummary struct {
	TotalMembers    int     `json:"totalMembers"`
	TotalHours      float64 `json:"totalHours"`
	AvgLoad         float64 `json:"avgLoad"`
	OverloadedCount int     `json:"overloadedCount"` // 利用率>100%的人数
}

type PMODashboardService struct{}

// GetPMODashboard 跨项目组合视图 + 健康度 RAG + 资源负荷汇总
func (s *PMODashboardService) GetPMODashboard() (*PMODashboard, error) {
	db := global.GVA_DB

	// 查询所有项目（eps_node 顶层）
	var projects []pmocker.PMEntity
	db.Where("entity_type = ? AND status != ?", "eps_node", "archived").Find(&projects)

	dash := &PMODashboard{
		HealthDist:   map[string]int{"green": 0, "yellow": 0, "red": 0},
		ProjectCards: []ProjectCard{},
	}

	ds := &DashboardService{}
	totalMembers := 0
	totalHours := 0.0
	overloaded := 0

	for _, p := range projects {
		card, err := s.GetProjectCard(p.ID)
		if err != nil {
			continue
		}
		dash.ProjectCards = append(dash.ProjectCards, *card)
		dash.HealthDist[card.Health]++

		// 资源负荷汇总
		rs := ds.calcResourceSummary(db, p.ID)
		totalMembers += rs.MemberCount
		totalHours += rs.TotalHours
		if rs.AvgUtilization > 100 {
			overloaded++
		}
	}

	dash.TotalProjects = len(dash.ProjectCards)
	dash.LoadSummary = LoadSummary{
		TotalMembers:    totalMembers,
		TotalHours:      totalHours,
		AvgLoad:         0,
		OverloadedCount: overloaded,
	}
	if totalMembers > 0 {
		dash.LoadSummary.AvgLoad = totalHours / float64(totalMembers) / 160.0 * 100
	}
	return dash, nil
}

// GetProjectHealth 基于进度偏差、成本偏差、风险数计算红黄绿
func (s *PMODashboardService) GetProjectHealth(projectID uint) string {
	db := global.GVA_DB
	ds := &DashboardService{}
	progress := ds.calcProgressByCount(db, projectID)
	cs := ds.calcCostSummary(db, projectID)
	rs := ds.calcRiskSummary(db, projectID)

	// 成本偏差率（正数=超支）
	costVarPct := 0.0
	if cs.Budget > 0 {
		costVarPct = (cs.Actual - cs.Budget) / cs.Budget * 100
	}
	// 进度偏差：简单以完成率衡量（<50% 且有近到期任务视为偏差）
	progressVarPct := 100 - progress

	if costVarPct > 20 || progressVarPct > 50 || rs.High >= 3 {
		return "red"
	}
	if costVarPct > 10 || progressVarPct > 30 || rs.High >= 1 {
		return "yellow"
	}
	return "green"
}

// GetProjectCard 返回项目卡片数据
func (s *PMODashboardService) GetProjectCard(projectID uint) (*ProjectCard, error) {
	db := global.GVA_DB
	var project pmocker.PMEntity
	if err := db.First(&project, projectID).Error; err != nil {
		return nil, err
	}
	ds := &DashboardService{}
	card := &ProjectCard{
		ProjectID:   project.ID,
		ProjectName: project.Title,
		Priority:    project.Priority,
		Status:      project.Status,
		Health:      s.GetProjectHealth(projectID),
		Progress:    ds.calcProgressByCount(db, projectID),
	}
	cs := ds.calcCostSummary(db, projectID)
	card.Budget = cs.Budget
	card.Actual = cs.Actual
	card.CostVariance = cs.Variance
	rs := ds.calcRiskSummary(db, projectID)
	card.RiskCount = rs.Open

	// 负责人姓名
	if project.OwnerID != nil {
		var user system.SysUser
		db.First(&user, *project.OwnerID)
		card.LeaderName = user.NickName
	}
	// 部门（从 EAV attr dept_id 读取，若未存则 0）
	card.DeptID = uint(getAttrInt(db, projectID, "dept_id"))
	return card, nil
}

// ===== 可见性 scope helper（Task 4/5 共用）=====

// VisibilityScope 用户可见性范围
type VisibilityScope struct {
	UserID        uint
	IsPMOAdmin    bool
	IsDeptLeader  bool
	IsCCBMember   bool
	DeptAncestors string // 部门负责人所在部门的物化路径前缀
	DeptID        uint
}

// getUserVisibilityScope 解析用户的可见性范围
// PMO_ADMIN(9001): 可见所有 P0/P1
// DEPT_LEADER 岗位: 可见本部门及子级下 P0/P1
// CCB_MEMBER 岗位: 可见所有 P0/P1 变更相关任务
// 其他: 仅见自己负责的 P0/P1
func getUserVisibilityScope(db *gorm.DB, userID uint) VisibilityScope {
	scope := VisibilityScope{UserID: userID}

	// 1. 检查 PMO_ADMIN 角色
	var authCount int64
	db.Table("sys_user_authority").
		Where("sys_user_id = ? AND sys_authority_authority_id = ?", userID, 9001).
		Count(&authCount)
	scope.IsPMOAdmin = authCount > 0

	// 2. 检查岗位
	var posCodes []string
	db.Table("sys_user_positions sup").
		Joins("JOIN sys_positions sp ON sp.id = sup.position_id").
		Where("sup.user_id = ?", userID).
		Pluck("sp.code", &posCodes)
	for _, code := range posCodes {
		if code == "DEPT_LEADER" {
			scope.IsDeptLeader = true
		}
		if code == "CCB_MEMBER" {
			scope.IsCCBMember = true
		}
	}

	// 3. 获取用户主部门及物化路径
	var user system.SysUser
	db.First(&user, userID)
	scope.DeptID = user.DeptId
	if scope.IsDeptLeader && user.DeptId > 0 {
		var dept system.SysDepartment
		db.First(&dept, user.DeptId)
		// ancestors 形如 "0,1,5"，用 LIKE 前缀匹配子级
		scope.DeptAncestors = fmt.Sprintf("%s,%d", dept.Ancestors, dept.ID)
	}
	return scope
}
