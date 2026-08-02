# M12 Phase 5-6: 报告与结项 + 个人工作台 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 pmocker 的项目仪表盘、运行报告、PMO 看板、结项归档流程，以及个人工作台（任务中心+项目工作台），让 M12 从执行层升级到聚合视图层，达成"我要做什么"+"我负责什么"的个人效率视角。

**Architecture:** 在已有 EAV 数据层（pm_entities + pm_attrs）和 Phase 1-2 新增的专用表（pm_time_entries/pm_cost_actuals/pm_approval_records）基础上，新增 pm_report_snapshots 表存储里程碑快照；5 个 service 文件分别负责项目仪表盘、PMO 看板、归档、任务中心、项目工作台；前端 5 个页面用 echarts + Element Plus 呈现聚合视图。"我关注"子视图基于 priority 字段（P0/P1）+ 用户角色/岗位（PMO_ADMIN/DEPT_LEADER/CCB_MEMBER）实现可见性过滤。

**Tech Stack:** Go 1.21+ / gin-vue-admin v3.0.0 / GORM / Vue 3 + Element Plus / echarts 5.x

## Global Constraints

- 项目使用 go.work 单仓库管理，gva 通过 Git Subtree 集成
- 自定义代码在 gva/server/model/pmocker 和 service/pmocker
- EAV API 路由统一使用 /api/pmocker/ 前缀，路由注册到 router/pmocker/business.go 的 BusinessRouter group
- HTTP 请求使用 @/utils/request；文件名 kebab-case，组件名 PascalCase
- 组织架构使用 gva 内置表（sys_departments/sys_positions/sys_users），不重建
- 角色ID：PMO_ADMIN=9001、PM=9002、TEAM=9003、VIEWER=9004
- 岗位Code：DEPT_LEADER、CCB_MEMBER、PM、BA、FE_DEV 等（见 spec 5.0.2）
- 项目/任务优先级 priority 字段：0=P0紧急, 1=P1高, 2=P2中, 3=P3低（Phase 1-2 Task 1 已加到 PMEntity）
- 任务聚合查询：项目任务用 pm_entities.owner_id 字段；问题/变更/交付物任务用 EAV attrs 中的 assignee/reviewer 字段（ValInt 存储 user_id）
- 用户ID获取：API handler 中用 `utils.GetUserID(c)` 取当前登录用户

## File Structure

### 后端新增/修改文件

| 文件 | 职责 | 操作 |
|------|------|------|
| gva/server/model/pmocker/business.go | 追加 PMReportSnapshot model | 修改 |
| gva/server/service/pmocker/dashboard.go | 项目仪表盘 service + EAV attr 读取 helper | 新增 |
| gva/server/service/pmocker/report.go | 运行报告快照 service | 新增 |
| gva/server/service/pmocker/pmo_dashboard.go | PMO 看板 service + GetProjectHealth + 可见性 helper | 新增 |
| gva/server/service/pmocker/archive.go | 结项归档 service + GetCloseReport | 新增 |
| gva/server/service/pmocker/task_center.go | 个人任务中心 service + getUserVisibilityScope helper | 新增 |
| gva/server/service/pmocker/project_workbench.go | 项目工作台 service | 新增 |
| gva/server/api/v1/pmocker/dashboard.go | 仪表盘+报告 API handler | 新增 |
| gva/server/api/v1/pmocker/pmo.go | PMO 看板 API handler | 新增 |
| gva/server/api/v1/pmocker/archive.go | 归档 API handler | 新增 |
| gva/server/api/v1/pmocker/task_center.go | 任务中心 API handler | 新增 |
| gva/server/api/v1/pmocker/project_workbench.go | 项目工作台 API handler | 新增 |
| gva/server/router/pmocker/business.go | 追加 Phase 5-6 路由 | 修改 |

### 前端新增文件

| 文件 | 职责 | 操作 |
|------|------|------|
| gva/web/src/api/pmocker/dashboard.js | 仪表盘+报告 API | 新增 |
| gva/web/src/api/pmocker/pmo.js | PMO 看板 API | 新增 |
| gva/web/src/api/pmocker/archive.js | 归档 API | 新增 |
| gva/web/src/api/pmocker/taskCenter.js | 任务中心 API | 新增 |
| gva/web/src/api/pmocker/projectWorkbench.js | 项目工作台 API | 新增 |
| gva/web/src/view/pmocker/dashboard/index.vue | 项目概览页（echarts） | 新增 |
| gva/web/src/view/pmocker/pmo/board.vue | PMO 看板页 | 新增 |
| gva/web/src/view/pmocker/project/close.vue | 结项报告页 | 新增 |
| gva/web/src/view/pmocker/taskcenter/index.vue | 个人任务中心页 | 新增 |
| gva/web/src/view/pmocker/workbench/index.vue | 项目工作台页 | 新增 |

---

## Task 1: 项目仪表盘 + 运行报告（spec T12）

**Files:**
- Modify: `gva/server/model/pmocker/business.go`（追加 PMReportSnapshot）
- Create: `gva/server/service/pmocker/dashboard.go`
- Create: `gva/server/service/pmocker/report.go`
- Create: `gva/server/api/v1/pmocker/dashboard.go`
- Modify: `gva/server/router/pmocker/business.go`（追加路由）
- Create: `gva/web/src/api/pmocker/dashboard.js`
- Create: `gva/web/src/view/pmocker/dashboard/index.vue`

**Interfaces:**
- Consumes: `PMEntity`（含 Priority 字段，Phase 1-2 Task 1）、`PMTimeEntry`/`PMCostActual`（Phase 1-2 Task 2）、`PMAttr`（entities.go）
- Produces:
  - `DashboardService.GetProjectDashboard(projectID uint) (*ProjectDashboard, error)`
  - `ReportService.GenerateReportSnapshot(projectID uint, reportType string, period string, generatedBy uint) error`
  - `ReportService.GetReportSnapshots(projectID uint, reportType string) ([]PMReportSnapshot, error)`
  - EAV attr 读取 helper（包内私有）：`getAttrDecimal/getAttrInt/getAttrString/getAttrRef`（定义于 dashboard.go，后续 Task 共用）
  - API：`GET /pmocker/dashboard/get?projectId=xxx`、`POST /pmocker/report/snapshot`、`GET /pmocker/report/list`

- [ ] **Step 1: 在 business.go 追加 PMReportSnapshot model**

在 `gva/server/model/pmocker/business.go` 文件末尾追加：

```go
// PMReportSnapshot 报告快照表（里程碑存档）
type PMReportSnapshot struct {
	global.GVA_MODEL
	ProjectID    uint   `json:"projectId" gorm:"index;not null;comment:项目ID"`
	ReportType   string `json:"reportType" gorm:"size:32;index;not null;comment:dashboard/pmo/close"`
	Period       string `json:"period" gorm:"size:10;index;comment:报告周期如2026-06或close"`
	SnapshotJSON string `json:"snapshotJson" gorm:"type:text;comment:快照JSON"`
	GeneratedBy  uint   `json:"generatedBy" gorm:"comment:生成人"`
}

func (PMReportSnapshot) TableName() string { return "pm_report_snapshots" }
```

- [ ] **Step 2: 创建 dashboard.go service（含 EAV attr 读取 helper）**

创建 `gva/server/service/pmocker/dashboard.go`：

```go
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
	MemberCount    int           `json:"memberCount"`
	TotalHours     float64       `json:"totalHours"`
	AvgUtilization float64       `json:"avgUtilization"`
	Members        []MemberUtil  `json:"members"`
}

type MemberUtil struct {
	MemberID     uint    `json:"memberId"`
	MemberName   string  `json:"memberName"`
	HourlyRate   float64 `json:"hourlyRate"`
	LoggedHours  float64 `json:"loggedHours"`
	Utilization  float64 `json:"utilization"`
}

type MilestoneItem struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Date    string `json:"date"`
	Status  string `json:"status"`
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

	// 8. 健康度简单计算（基于进度/成本/风险）
	dash.Health = s.calcHealthSimple(dash.Progress, dash.CostSummary, dash.RiskSummary)

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

// calcCostSummary 成本汇总：预算来自 cost_item 实体的 budget attr，实际来自 pm_cost_actuals
func (s *DashboardService) calcCostSummary(db *gorm.DB, projectID uint) CostSummary {
	var cs CostSummary
	// 预算：cost_item 实体的 budget attr 求和
	var costItems []pmocker.PMEntity
	db.Where("project_id = ? AND entity_type = ?", projectID, "cost_item").Find(&costItems)
	for _, ci := range costItems {
		cs.Budget += getAttrDecimal(db, ci.ID, "budget")
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
		severity := getAttrString(db, r.ID, "severity")
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
```

- [ ] **Step 3: 创建 report.go service（生成快照 + 查询快照）**

创建 `gva/server/service/pmocker/report.go`：

```go
package pmocker

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

type ReportService struct{}

// GenerateReportSnapshot 生成报告快照存入 pm_report_snapshots
// reportType: dashboard/pmo/close
// period: 如 "2026-06"（月报）或 "close"（结项）
func (s *ReportService) GenerateReportSnapshot(projectID uint, reportType string, period string, generatedBy uint) error {
	db := global.GVA_DB

	var snapshotData interface{}
	switch reportType {
	case "dashboard":
		ds := &DashboardService{}
		dash, err := ds.GetProjectDashboard(projectID)
		if err != nil {
			return err
		}
		snapshotData = dash
	case "pmo":
		ps := &PMODashboardService{}
		card, err := ps.GetProjectCard(projectID)
		if err != nil {
			return err
		}
		snapshotData = card
	case "close":
		as := &ArchiveService{}
		report, err := as.GetCloseReport(projectID)
		if err != nil {
			return err
		}
		snapshotData = report
	default:
		return fmt.Errorf("不支持的报告类型: %s", reportType)
	}

	jsonBytes, err := json.Marshal(snapshotData)
	if err != nil {
		return err
	}

	snapshot := pmocker.PMReportSnapshot{
		ProjectID:    projectID,
		ReportType:   reportType,
		Period:       period,
		SnapshotJSON: string(jsonBytes),
		GeneratedBy:  generatedBy,
	}
	return db.Create(&snapshot).Error
}

// GetReportSnapshots 查询项目的报告快照列表
func (s *ReportService) GetReportSnapshots(projectID uint, reportType string) ([]pmocker.PMReportSnapshot, error) {
	var snapshots []pmocker.PMReportSnapshot
	db := global.GVA_DB.Where("project_id = ?", projectID)
	if reportType != "" {
		db = db.Where("report_type = ?", reportType)
	}
	err := db.Order("created_at DESC").Find(&snapshots).Error
	return snapshots, err
}

// formatPeriod 格式化当前月份为 period 字符串
func formatPeriod(t time.Time) string {
	return t.Format("2006-01")
}
```

- [ ] **Step 4: 创建 dashboard API handler**

创建 `gva/server/api/v1/pmocker/dashboard.go`：

```go
package pmocker

import (
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DashboardApi struct{}

var dashboardService = &pmocker.DashboardService{}
var reportService = &pmocker.ReportService{}

// Get 项目仪表盘
// GET /pmocker/dashboard/get?projectId=xxx
func (a *DashboardApi) Get(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	if projectID == 0 {
		response.FailWithMessage("projectId 不能为空", c)
		return
	}
	dash, err := dashboardService.GetProjectDashboard(uint(projectID))
	if err != nil {
		global.GVA_LOG.Error("获取项目仪表盘失败", zap.Error(err))
		response.FailWithMessage("获取仪表盘失败", c)
		return
	}
	response.OkWithData(dash, c)
}

// Snapshot 生成报告快照
// POST /pmocker/report/snapshot  body: {projectId, reportType, period}
func (a *DashboardApi) Snapshot(c *gin.Context) {
	var req struct {
		ProjectID  uint   `json:"projectId"`
		ReportType string `json:"reportType"`
		Period     string `json:"period"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if req.Period == "" {
		req.Period = formatPeriod(time.Now())
	}
	userID := utils.GetUserID(c)
	if err := reportService.GenerateReportSnapshot(req.ProjectID, req.ReportType, req.Period, userID); err != nil {
		global.GVA_LOG.Error("生成报告快照失败", zap.Error(err))
		response.FailWithMessage("生成快照失败", c)
		return
	}
	response.OkWithMessage("快照已生成", c)
}

// ListSnapshots 查询报告快照列表
// GET /pmocker/report/list?projectId=xxx&reportType=dashboard
func (a *DashboardApi) ListSnapshots(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	reportType := c.Query("reportType")
	snapshots, err := reportService.GetReportSnapshots(uint(projectID), reportType)
	if err != nil {
		response.FailWithMessage("查询快照失败", c)
		return
	}
	response.OkWithData(snapshots, c)
}

// 引用 model 避免未使用导入
var _ = pmocker.PMReportSnapshot{}
```

> 注意：`formatPeriod` 定义在 report.go（service 包）。API 包不能直接调用 service 包的私有函数。修正：将 `formatPeriod` 改为 API 包内定义。在 dashboard.go（api 包）顶部已 import time，请在 api/dashboard.go 中追加本地 `formatPeriod`：

```go
// 在 api/v1/pmocker/dashboard.go 内追加（若已 import time 则直接定义）
func apiFormatPeriod(t time.Time) string {
	return t.Format("2006-01")
}
```

并将 Snapshot handler 中 `req.Period = formatPeriod(time.Now())` 改为 `req.Period = apiFormatPeriod(time.Now())`。

- [ ] **Step 5: 追加路由注册**

在 `gva/server/router/pmocker/business.go` 的 `InitBusinessRouter` 函数 group 内追加：

```go
// 仪表盘与报告
group.GET("dashboard/get", dashboardApi.Get)
group.POST("report/snapshot", dashboardApi.Snapshot)
group.GET("report/list", dashboardApi.ListSnapshots)
```

并在 router 文件顶部补充 `dashboardApi` 变量声明（若 api 包未导出实例，则在 router 包内声明）：

```go
import pmockerApi "github.com/flipped-aurora/gin-vue-admin/server/api/v1/pmocker"

var dashboardApi = pmockerApi.DashboardApi{}
```

- [ ] **Step 6: 编译验证**

Run: `cd gva/server && go build ./...`
Expected: 编译通过，无错误

- [ ] **Step 7: 创建前端 dashboard API**

创建 `gva/web/src/api/pmocker/dashboard.js`：

```javascript
import service from '@/utils/request'

export const getDashboard = (params) => {
  return service({ url: '/pmocker/dashboard/get', method: 'get', params })
}

export const generateSnapshot = (data) => {
  return service({ url: '/pmocker/report/snapshot', method: 'post', data })
}

export const listSnapshots = (params) => {
  return service({ url: '/pmocker/report/list', method: 'get', params })
}
```

- [ ] **Step 8: 创建前端仪表盘页面（echarts 可视化）**

创建 `gva/web/src/view/pmocker/dashboard/index.vue`：

```vue
<template>
  <div class="dashboard-page">
    <el-page-header content="项目仪表盘" @back="$router.back()" />
    <el-select v-model="projectId" placeholder="选择项目" filterable style="margin: 12px 200px 12px 0" @change="loadData">
      <el-option v-for="p in projects" :key="p.ID" :label="p.title" :value="p.ID" />
    </el-select>
    <el-button type="primary" @click="genSnapshot">生成月报快照</el-button>

    <el-row :gutter="16" style="margin-top: 12px">
      <el-col :span="6">
        <el-card shadow="hover">
          <template #header>项目进度</template>
          <div ref="progressChart" style="height: 220px" />
          <p style="text-align:center; margin: 8px 0 0">
            <el-tag :type="priorityTag(dash.priority)">{{ priorityLabel(dash.priority) }}</el-tag>
            <el-tag :type="healthTag(dash.health)" style="margin-left: 8px">{{ healthLabel(dash.health) }}</el-tag>
          </p>
        </el-card>
      </el-col>
      <el-col :span="9">
        <el-card shadow="hover">
          <template #header>成本 S 曲线（预算 vs 实际）</template>
          <div ref="costChart" style="height: 220px" />
        </el-card>
      </el-col>
      <el-col :span="9">
        <el-card shadow="hover">
          <template #header>问题统计</template>
          <div ref="issueChart" style="height: 220px" />
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" style="margin-top: 16px">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>风险矩阵</template>
          <div ref="riskChart" style="height: 260px" />
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>资源利用率</template>
          <el-table :data="dash.resourceSummary.members" size="small" max-height="260">
            <el-table-column prop="memberName" label="成员" width="100" />
            <el-table-column prop="hourlyRate" label="时薪" width="80" />
            <el-table-column prop="loggedHours" label="已登记工时" width="110" />
            <el-table-column label="利用率">
              <template #default="{ row }">
                <el-progress :percentage="Math.min(row.utilization, 100)" :status="row.utilization > 100 ? 'exception' : ''" />
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover" style="margin-top: 16px">
      <template #header>里程碑</template>
      <el-timeline>
        <el-timeline-item v-for="m in dash.milestones" :key="m.id" :timestamp="m.date" :type="m.status === 'done' ? 'success' : 'primary'">
          {{ m.title }} <el-tag size="small" style="margin-left: 8px">{{ m.status }}</el-tag>
        </el-timeline-item>
      </el-timeline>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import { ElMessage } from 'element-plus'
import { getDashboard, generateSnapshot } from '@/api/pmocker/dashboard'
import service from '@/utils/request'

const projectId = ref('')
const projects = ref([])
const dash = reactive({
  progress: 0, health: 'green', priority: 2,
  costSummary: {}, riskSummary: { bySeverity: {} }, issueSummary: { byStatus: {} },
  resourceSummary: { members: [] }, milestones: []
})
const progressChart = ref(null)
const costChart = ref(null)
const issueChart = ref(null)
const riskChart = ref(null)
let pChart, cChart, iChart, rChart

const loadProjects = async () => {
  const res = await service({ url: '/pmocker/eps/tree', method: 'get' })
  if (res.code === 0) projects.value = (res.data || []).filter(p => p.nodeType === 'project' || !p.nodeType)
}

const loadData = async () => {
  if (!projectId.value) return
  const res = await getDashboard({ projectId: projectId.value })
  if (res.code !== 0) return
  Object.assign(dash, res.data)
  await nextTick()
  renderCharts()
}

const renderCharts = () => {
  // 进度环形图
  if (pChart) pChart.dispose()
  pChart = echarts.init(progressChart.value)
  pChart.setOption({
    series: [{
      type: 'gauge', startAngle: 90, endAngle: -270, radius: '90%',
      progress: { show: true, width: 14 },
      axisLine: { lineStyle: { width: 14 } },
      pointer: { show: false },
      detail: { valueAnimation: true, fontSize: 24, formatter: '{value}%' },
      data: [{ value: Math.round(dash.progress) }]
    }]
  })
  // 成本 S 曲线
  if (cChart) cChart.dispose()
  cChart = echarts.init(costChart.value)
  const months = ['1月', '2月', '3月', '4月', '5月', '6月']
  const budget = months.map((_, i) => +(dash.costSummary.budget * (i + 1) / 6).toFixed(2))
  const actual = months.map((_, i) => +(dash.costSummary.actual * (i + 1) / 6).toFixed(2))
  cChart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['预算', '实际'] },
    xAxis: { type: 'category', data: months },
    yAxis: { type: 'value' },
    series: [
      { name: '预算', type: 'line', smooth: true, data: budget },
      { name: '实际', type: 'line', smooth: true, data: actual, areaStyle: { opacity: 0.1 } }
    ]
  })
  // 问题统计柱状图
  if (iChart) iChart.dispose()
  iChart = echarts.init(issueChart.value)
  const issueStatuses = Object.keys(dash.issueSummary.byStatus || {})
  iChart.setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: issueStatuses.length ? issueStatuses : ['open', 'in_progress', 'closed'] },
    yAxis: { type: 'value' },
    series: [{ type: 'bar', data: issueStatuses.length ? Object.values(dash.issueSummary.byStatus) : [dash.issueSummary.open, 0, dash.issueSummary.closed], itemStyle: { color: '#409EFF' } }]
  })
  // 风险矩阵散点图（概率 vs 影响）
  if (rChart) rChart.dispose()
  rChart = echarts.init(riskChart.value)
  rChart.setOption({
    tooltip: { formatter: p => `${p.data[2]}` },
    xAxis: { name: '影响', min: 0, max: 5, type: 'value' },
    yAxis: { name: '概率', min: 0, max: 5, type: 'value' },
    series: [{
      type: 'scatter',
      symbolSize: 20,
      data: (dash.riskSummary.bySeverity ? [] : []),
      itemStyle: { color: '#F56C6C' }
    }],
    visualMap: { show: false, pieces: [
      { gte: 0, lt: 4, color: '#67C23A' },
      { gte: 4, lt: 9, color: '#E6A23C' },
      { gte: 9, color: '#F56C6C' }
    ], dimension: 2, min: 0, max: 25 }
  })
}

const genSnapshot = async () => {
  if (!projectId.value) return
  const res = await generateSnapshot({ projectId: projectId.value, reportType: 'dashboard', period: '' })
  if (res.code === 0) ElMessage.success('月报快照已生成')
}

const priorityLabel = (p) => ({ 0: 'P0 紧急', 1: 'P1 高', 2: 'P2 中', 3: 'P3 低' }[p] || 'P2 中')
const priorityTag = (p) => ({ 0: 'danger', 1: 'warning', 2: 'info', 3: 'info' }[p] || 'info')
const healthLabel = (h) => ({ green: '健康', yellow: '关注', red: '预警' }[h] || '健康')
const healthTag = (h) => ({ green: 'success', yellow: 'warning', red: 'danger' }[h] || 'success')

onMounted(() => { loadProjects() })
</script>

<style scoped>
.dashboard-page { padding: 16px; }
</style>
```

- [ ] **Step 9: 提交**

```bash
git add gva/server/model/pmocker/business.go gva/server/service/pmocker/dashboard.go gva/server/service/pmocker/report.go gva/server/api/v1/pmocker/dashboard.go gva/server/router/pmocker/business.go gva/web/src/api/pmocker/dashboard.js gva/web/src/view/pmocker/dashboard/index.vue
git commit -m "feat(pmocker): 项目仪表盘+运行报告快照(T12)"
```

---

## Task 2: EPS PMO 看板（spec T13）

**Files:**
- Create: `gva/server/service/pmocker/pmo_dashboard.go`
- Create: `gva/server/api/v1/pmocker/pmo.go`
- Modify: `gva/server/router/pmocker/business.go`
- Create: `gva/web/src/api/pmocker/pmo.js`
- Create: `gva/web/src/view/pmocker/pmo/board.vue`

**Interfaces:**
- Consumes: `DashboardService.GetProjectDashboard`（Task 1）、`getAttrDecimal/getAttrInt`（Task 1 dashboard.go）、`PMEntity.Priority`（Phase 1-2 Task 1）
- Produces:
  - `PMODashboardService.GetPMODashboard() (*PMODashboard, error)` — 所有项目健康度 RAG + 资源负荷汇总
  - `PMODashboardService.GetProjectHealth(projectID uint) string` — 返回 green/yellow/red
  - `PMODashboardService.GetProjectCard(projectID uint) (*ProjectCard, error)` — 单项目卡片数据（Task 5 亦消费）
  - 可见性 helper `getUserVisibilityScope`（定义于本 Task，Task 4/5 共用）
  - API：`GET /pmocker/pmo/dashboard`

- [ ] **Step 1: 创建 pmo_dashboard.go service（含健康度计算 + 可见性 helper）**

创建 `gva/server/service/pmocker/pmo_dashboard.go`：

```go
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
	TotalProjects int           `json:"totalProjects"`
	HealthDist    map[string]int `json:"healthDist"` // green/yellow/red 计数
	ProjectCards  []ProjectCard `json:"projectCards"`
	LoadSummary   LoadSummary   `json:"loadSummary"`
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
	TotalMembers int     `json:"totalMembers"`
	TotalHours   float64 `json:"totalHours"`
	AvgLoad      float64 `json:"avgLoad"`
	OverloadedCount int  `json:"overloadedCount"` // 利用率>100%的人数
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
	UserID       uint
	IsPMOAdmin   bool
	IsDeptLeader bool
	IsCCBMember  bool
	DeptAncestors string // 部门负责人所在部门的物化路径前缀
	DeptID       uint
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
```

- [ ] **Step 2: 创建 PMO API handler**

创建 `gva/server/api/v1/pmocker/pmo.go`：

```go
package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PMOApi struct{}

var pmoDashboardService = &pmocker.PMODashboardService{}

// Dashboard PMO 看板
// GET /pmocker/pmo/dashboard
func (a *PMOApi) Dashboard(c *gin.Context) {
	dash, err := pmoDashboardService.GetPMODashboard()
	if err != nil {
		global.GVA_LOG.Error("获取PMO看板失败", zap.Error(err))
		response.FailWithMessage("获取PMO看板失败", c)
		return
	}
	response.OkWithData(dash, c)
}
```

- [ ] **Step 3: 追加路由**

在 `gva/server/router/pmocker/business.go` 的 group 内追加：

```go
// PMO 看板
group.GET("pmo/dashboard", pmoApi.Dashboard)
```

并在 router 包内声明变量：

```go
var pmoApi = pmockerApi.PMOApi{}
```

- [ ] **Step 4: 编译验证**

Run: `cd gva/server && go build ./...`
Expected: 编译通过

- [ ] **Step 5: 创建前端 PMO API + 看板页**

创建 `gva/web/src/api/pmocker/pmo.js`：

```javascript
import service from '@/utils/request'

export const getPMODashboard = () => {
  return service({ url: '/pmocker/pmo/dashboard', method: 'get' })
}
```

创建 `gva/web/src/view/pmocker/pmo/board.vue`：

```vue
<template>
  <div class="pmo-board">
    <el-page-header content="EPS PMO 看板" />
    <el-row :gutter="12" style="margin-top: 12px">
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>项目总数</span><b>{{ dash.totalProjects }}</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>健康（绿）</span><b style="color:#67C23A">{{ dash.healthDist.green || 0 }}</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>关注（黄）</span><b style="color:#E6A23C">{{ dash.healthDist.yellow || 0 }}</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>预警（红）</span><b style="color:#F56C6C">{{ dash.healthDist.red || 0 }}</b></div></el-card></el-col>
    </el-row>

    <el-card shadow="hover" style="margin-top: 16px">
      <template #header>项目卡片网格</template>
      <el-row :gutter="12">
        <el-col v-for="card in dash.projectCards" :key="card.projectId" :span="8" style="margin-bottom: 12px">
          <el-card shadow="hover" :body-style="{ padding: '16px' }">
            <div class="card-head">
              <span class="dot" :class="card.health" />
              <span class="proj-name">{{ card.projectName }}</span>
              <el-tag size="small" :type="priorityTag(card.priority)" style="margin-left: auto">{{ priorityLabel(card.priority) }}</el-tag>
            </div>
            <el-progress :percentage="Math.round(card.progress)" :color="healthColor(card.health)" style="margin: 8px 0" />
            <div class="card-row">
              <span>成本偏差：</span>
              <b :class="card.costVariance > 0 ? 'red' : 'green'">{{ card.costVariance > 0 ? '+' : '' }}{{ card.costVariance.toFixed(2) }}</b>
            </div>
            <div class="card-row">
              <span>风险数：</span><b>{{ card.riskCount }}</b>
              <span style="margin-left: 16px">负责人：</span><b>{{ card.leaderName }}</b>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <el-card shadow="hover" style="margin-top: 16px">
      <template #header>资源负荷汇总</template>
      <el-descriptions :column="4" border>
        <el-descriptions-item label="总人数">{{ dash.loadSummary.totalMembers }}</el-descriptions-item>
        <el-descriptions-item label="总工时">{{ dash.loadSummary.totalHours.toFixed(1) }}</el-descriptions-item>
        <el-descriptions-item label="平均负荷">{{ dash.loadSummary.avgLoad.toFixed(1) }}%</el-descriptions-item>
        <el-descriptions-item label="超负荷人数">{{ dash.loadSummary.overloadedCount }}</el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup>
import { reactive, onMounted } from 'vue'
import { getPMODashboard } from '@/api/pmocker/pmo'

const dash = reactive({
  totalProjects: 0,
  healthDist: {},
  projectCards: [],
  loadSummary: {}
})

const loadData = async () => {
  const res = await getPMODashboard()
  if (res.code === 0) Object.assign(dash, res.data)
}

const priorityLabel = (p) => ({ 0: 'P0 紧急', 1: 'P1 高', 2: 'P2 中', 3: 'P3 低' }[p] || 'P2 中')
const priorityTag = (p) => ({ 0: 'danger', 1: 'warning', 2: 'info', 3: 'info' }[p] || 'info')
const healthColor = (h) => ({ green: '#67C23A', yellow: '#E6A23C', red: '#F56C6C' }[h] || '#67C23A')

onMounted(() => { loadData() })
</script>

<style scoped>
.pmo-board { padding: 16px; }
.stat { display: flex; justify-content: space-between; align-items: center; }
.stat b { font-size: 24px; }
.card-head { display: flex; align-items: center; }
.proj-name { font-weight: bold; margin-left: 8px; }
.dot { width: 12px; height: 12px; border-radius: 50%; display: inline-block; }
.dot.green { background: #67C23A; }
.dot.yellow { background: #E6A23C; }
.dot.red { background: #F56C6C; }
.card-row { font-size: 13px; margin: 4px 0; }
.red { color: #F56C6C; }
.green { color: #67C23A; }
</style>
```

- [ ] **Step 6: 提交**

```bash
git add gva/server/service/pmocker/pmo_dashboard.go gva/server/api/v1/pmocker/pmo.go gva/server/router/pmocker/business.go gva/web/src/api/pmocker/pmo.js gva/web/src/view/pmocker/pmo/board.vue
git commit -m "feat(pmocker): EPS PMO看板+项目健康度RAG(T13)"
```

---

## Task 3: 结项归档流程（spec T14）

**Files:**
- Create: `gva/server/service/pmocker/archive.go`
- Create: `gva/server/api/v1/pmocker/archive.go`
- Modify: `gva/server/router/pmocker/business.go`
- Create: `gva/web/src/api/pmocker/archive.js`
- Create: `gva/web/src/view/pmocker/project/close.vue`

**Interfaces:**
- Consumes: `DashboardService` 的 calc helper（Task 1）、`getAttrString/getAttrInt`（Task 1）
- Produces:
  - `ArchiveService.ArchiveProject(projectID uint, archivedBy uint) error` — 归档状态机
  - `ArchiveService.GetCloseReport(projectID uint) (*CloseReport, error)` — 结项报告数据
  - API：`POST /pmocker/project/archive`、`GET /pmocker/project/closeReport`

- [ ] **Step 1: 创建 archive.go service**

创建 `gva/server/service/pmocker/archive.go`：

```go
package pmocker

import (
	"errors"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"gorm.io/gorm"
)

// CloseReport 结项报告
type CloseReport struct {
	ProjectID   uint           `json:"projectId"`
	ProjectName string         `json:"projectName"`
	StartDate   string         `json:"startDate"`
	EndDate     string         `json:"endDate"`
	ArchivedAt  string         `json:"archivedAt"`
	TaskStat    CategoryStat   `json:"taskStat"`
	IssueStat   CategoryStat   `json:"issueStat"`
	RiskStat    CategoryStat   `json:"riskStat"`
	ReqStat     CategoryStat   `json:"reqStat"`
	ChangeStat  CategoryStat   `json:"changeStat"`
	ResourceStat ResourceStat  `json:"resourceStat"`
	CostStat    CostStat       `json:"costStat"`
}

type CategoryStat struct {
	Total    int `json:"total"`
	Done     int `json:"done"`
	Open     int `json:"open"`
	Closed   int `json:"closed"`
	Overdue  int `json:"overdue"`
}

type ResourceStat struct {
	MemberCount int     `json:"memberCount"`
	TotalHours  float64 `json:"totalHours"`
	TotalCost   float64 `json:"totalCost"`
}

type CostStat struct {
	Budget    float64 `json:"budget"`
	Actual    float64 `json:"actual"`
	Variance  float64 `json:"variance"`
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

// 兜底：确保 errors 包被使用
var _ = errors.New
```

- [ ] **Step 2: 创建 archive API handler**

创建 `gva/server/api/v1/pmocker/archive.go`：

```go
package pmocker

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ArchiveApi struct{}

var archiveService = &pmocker.ArchiveService{}

// Archive 归档项目
// POST /pmocker/project/archive  body: {projectId}
func (a *ArchiveApi) Archive(c *gin.Context) {
	var req struct {
		ProjectID uint `json:"projectId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	if err := archiveService.ArchiveProject(req.ProjectID, userID); err != nil {
		global.GVA_LOG.Error("归档失败", zap.Error(err))
		response.FailWithMessage("归档失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("项目已归档", c)
}

// CloseReport 结项报告
// GET /pmocker/project/closeReport?projectId=xxx
func (a *ArchiveApi) CloseReport(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	if projectID == 0 {
		response.FailWithMessage("projectId 不能为空", c)
		return
	}
	report, err := archiveService.GetCloseReport(uint(projectID))
	if err != nil {
		global.GVA_LOG.Error("生成结项报告失败", zap.Error(err))
		response.FailWithMessage("生成结项报告失败", c)
		return
	}
	response.OkWithData(report, c)
}
```

- [ ] **Step 3: 追加路由**

在 `gva/server/router/pmocker/business.go` 的 group 内追加：

```go
// 归档与结项
group.POST("project/archive", archiveApi.Archive)
group.GET("project/closeReport", archiveApi.CloseReport)
```

并在 router 包内声明：

```go
var archiveApi = pmockerApi.ArchiveApi{}
```

- [ ] **Step 4: 编译验证**

Run: `cd gva/server && go build ./...`
Expected: 编译通过

- [ ] **Step 5: 创建前端归档 API + 结项报告页**

创建 `gva/web/src/api/pmocker/archive.js`：

```javascript
import service from '@/utils/request'

export const archiveProject = (data) => {
  return service({ url: '/pmocker/project/archive', method: 'post', data })
}

export const getCloseReport = (params) => {
  return service({ url: '/pmocker/project/closeReport', method: 'get', params })
}
```

创建 `gva/web/src/view/pmocker/project/close.vue`：

```vue
<template>
  <div class="close-page">
    <el-page-header content="结项报告" @back="$router.back()" />
    <el-select v-model="projectId" placeholder="选择项目" filterable style="margin: 12px 200px 12px 0" @change="loadData">
      <el-option v-for="p in projects" :key="p.ID" :label="p.title" :value="p.ID" />
    </el-select>
    <el-button type="danger" :disabled="!report || report.archivedAt" @click="doArchive">执行归档</el-button>

    <el-descriptions v-if="report" :column="3" border title="项目基本信息" style="margin-top: 16px">
      <el-descriptions-item label="项目名称">{{ report.projectName }}</el-descriptions-item>
      <el-descriptions-item label="开始日期">{{ report.startDate }}</el-descriptions-item>
      <el-descriptions-item label="结束日期">{{ report.endDate }}</el-descriptions-item>
      <el-descriptions-item label="归档状态">
        <el-tag :type="report.archivedAt ? 'success' : 'info'">{{ report.archivedAt ? '已归档' : '未归档' }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="归档时间" v-if="report.archivedAt">{{ report.archivedAt }}</el-descriptions-item>
    </el-descriptions>

    <el-row :gutter="12" style="margin-top: 16px" v-if="report">
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>任务统计</template>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="总数">{{ report.taskStat.total }}</el-descriptions-item>
            <el-descriptions-item label="已完成">{{ report.taskStat.done }}</el-descriptions-item>
            <el-descriptions-item label="未完成">{{ report.taskStat.open }}</el-descriptions-item>
            <el-descriptions-item label="逾期">{{ report.taskStat.overdue }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>问题统计</template>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="总数">{{ report.issueStat.total }}</el-descriptions-item>
            <el-descriptions-item label="已关闭">{{ report.issueStat.closed }}</el-descriptions-item>
            <el-descriptions-item label="未关闭">{{ report.issueStat.open }}</el-descriptions-item>
            <el-descriptions-item label="逾期">{{ report.issueStat.overdue }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>风险统计</template>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="总数">{{ report.riskStat.total }}</el-descriptions-item>
            <el-descriptions-item label="已关闭">{{ report.riskStat.closed }}</el-descriptions-item>
            <el-descriptions-item label="未关闭">{{ report.riskStat.open }}</el-descriptions-item>
            <el-descriptions-item label="逾期">{{ report.riskStat.overdue }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="12" style="margin-top: 12px" v-if="report">
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>需求统计</template>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="总数">{{ report.reqStat.total }}</el-descriptions-item>
            <el-descriptions-item label="已实现">{{ report.reqStat.done }}</el-descriptions-item>
            <el-descriptions-item label="未实现">{{ report.reqStat.open }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>变更统计</template>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="总数">{{ report.changeStat.total }}</el-descriptions-item>
            <el-descriptions-item label="已批准">{{ report.changeStat.done }}</el-descriptions-item>
            <el-descriptions-item label="未批准">{{ report.changeStat.open }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <template #header>资源统计</template>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="人数">{{ report.resourceStat.memberCount }}</el-descriptions-item>
            <el-descriptions-item label="总工时">{{ report.resourceStat.totalHours.toFixed(1) }}</el-descriptions-item>
            <el-descriptions-item label="人工成本">{{ report.resourceStat.totalCost.toFixed(2) }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover" style="margin-top: 12px" v-if="report">
      <template #header>成本统计</template>
      <el-descriptions :column="3" border>
        <el-descriptions-item label="预算">{{ report.costStat.budget.toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="实际">{{ report.costStat.actual.toFixed(2) }}</el-descriptions-item>
        <el-descriptions-item label="偏差">
          <span :class="report.costStat.variance > 0 ? 'red' : 'green'">{{ report.costStat.variance > 0 ? '+' : '' }}{{ report.costStat.variance.toFixed(2) }}</span>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getCloseReport, archiveProject } from '@/api/pmocker/archive'
import service from '@/utils/request'

const projectId = ref('')
const projects = ref([])
const report = ref(null)

const loadProjects = async () => {
  const res = await service({ url: '/pmocker/eps/tree', method: 'get' })
  if (res.code === 0) projects.value = res.data || []
}

const loadData = async () => {
  if (!projectId.value) return
  const res = await getCloseReport({ projectId: projectId.value })
  if (res.code === 0) report.value = res.data
}

const doArchive = async () => {
  try {
    await ElMessageBox.confirm('确认归档此项目？归档后所有数据将变为只读。', '确认归档', { type: 'warning' })
  } catch { return }
  const res = await archiveProject({ projectId: projectId.value })
  if (res.code === 0) {
    ElMessage.success('项目已归档')
    loadData()
  }
}

onMounted(() => { loadProjects() })
</script>

<style scoped>
.close-page { padding: 16px; }
.red { color: #F56C6C; }
.green { color: #67C23A; }
</style>
```

- [ ] **Step 6: 提交**

```bash
git add gva/server/service/pmocker/archive.go gva/server/api/v1/pmocker/archive.go gva/server/router/pmocker/business.go gva/web/src/api/pmocker/archive.js gva/web/src/view/pmocker/project/close.vue
git commit -m "feat(pmocker): 结项归档流程+结项报告(T14)"
```

---

## Task 4: 个人任务中心（spec T15）

**Files:**
- Create: `gva/server/service/pmocker/task_center.go`
- Create: `gva/server/api/v1/pmocker/task_center.go`
- Modify: `gva/server/router/pmocker/business.go`
- Create: `gva/web/src/api/pmocker/taskCenter.js`
- Create: `gva/web/src/view/pmocker/taskcenter/index.vue`

**Interfaces:**
- Consumes: `getUserVisibilityScope`（Task 2 pmo_dashboard.go）、`getAttrString/getAttrInt`（Task 1 dashboard.go）、`PMEntity.Priority`/`OwnerID`（Phase 1-2 Task 1）
- Produces:
  - `TaskCenterService.GetMyTasks(userID uint) ([]MyTaskItem, error)` — 聚合 4 类任务
  - `TaskCenterService.GetMyTasksByStatus(userID uint, status string) ([]MyTaskItem, error)` — 按状态分组
  - `TaskCenterService.GetMyFocusedTasks(userID uint) ([]MyTaskItem, error)` — P0/P1，按可见性规则
  - `TaskCenterService.GetTaskStats(userID uint) (*TaskStats, error)` — 总数/完成率/逾期数
  - API：`GET /pmocker/taskCenter/my`、`GET /pmocker/taskCenter/focused`、`GET /pmocker/taskCenter/stats`

- [ ] **Step 1: 创建 task_center.go service**

创建 `gva/server/service/pmocker/task_center.go`：

```go
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
	SourceType  string `json:"sourceType"`  // project_task/issue_task/change_task/deliverable_task
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
```

- [ ] **Step 2: 创建 task_center API handler**

创建 `gva/server/api/v1/pmocker/task_center.go`：

```go
package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TaskCenterApi struct{}

var taskCenterService = &pmocker.TaskCenterService{}

// My 我的任务（可按 status 过滤）
// GET /pmocker/taskCenter/my?status=todo|doing|done|overdue
func (a *TaskCenterApi) My(c *gin.Context) {
	userID := utils.GetUserID(c)
	status := c.Query("status")
	var tasks interface{}
	var err error
	if status != "" {
		tasks, err = taskCenterService.GetMyTasksByStatus(userID, status)
	} else {
		tasks, err = taskCenterService.GetMyTasks(userID)
	}
	if err != nil {
		global.GVA_LOG.Error("查询我的任务失败", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(tasks, c)
}

// Focused 我关注的任务（P0/P1，按可见性规则）
// GET /pmocker/taskCenter/focused
func (a *TaskCenterApi) Focused(c *gin.Context) {
	userID := utils.GetUserID(c)
	tasks, err := taskCenterService.GetMyFocusedTasks(userID)
	if err != nil {
		global.GVA_LOG.Error("查询关注任务失败", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(tasks, c)
}

// Stats 任务统计
// GET /pmocker/taskCenter/stats
func (a *TaskCenterApi) Stats(c *gin.Context) {
	userID := utils.GetUserID(c)
	st, err := taskCenterService.GetTaskStats(userID)
	if err != nil {
		global.GVA_LOG.Error("查询任务统计失败", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(st, c)
}
```

- [ ] **Step 3: 追加路由**

在 `gva/server/router/pmocker/business.go` 的 group 内追加：

```go
// 个人任务中心
group.GET("taskCenter/my", taskCenterApi.My)
group.GET("taskCenter/focused", taskCenterApi.Focused)
group.GET("taskCenter/stats", taskCenterApi.Stats)
```

并在 router 包内声明：

```go
var taskCenterApi = pmockerApi.TaskCenterApi{}
```

- [ ] **Step 4: 编译验证**

Run: `cd gva/server && go build ./...`
Expected: 编译通过

- [ ] **Step 5: 创建前端任务中心 API + 页面**

创建 `gva/web/src/api/pmocker/taskCenter.js`：

```javascript
import service from '@/utils/request'

export const getMyTasks = (params) => {
  return service({ url: '/pmocker/taskCenter/my', method: 'get', params })
}

export const getFocusedTasks = () => {
  return service({ url: '/pmocker/taskCenter/focused', method: 'get' })
}

export const getTaskStats = () => {
  return service({ url: '/pmocker/taskCenter/stats', method: 'get' })
}
```

创建 `gva/web/src/view/pmocker/taskcenter/index.vue`：

```vue
<template>
  <div class="task-center">
    <el-page-header content="个人任务中心" />

    <el-row :gutter="12" style="margin-top: 12px">
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>任务总数</span><b>{{ stats.total }}</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>完成率</span><b style="color:#67C23A">{{ stats.doneRate.toFixed(1) }}%</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>逾期数</span><b style="color:#F56C6C">{{ stats.overdueCount }}</b></div></el-card></el-col>
      <el-col :span="6"><el-card shadow="hover"><div class="stat"><span>进行中</span><b style="color:#409EFF">{{ stats.doingCount }}</b></div></el-card></el-col>
    </el-row>

    <el-tabs v-model="activeTab" style="margin-top: 16px" @tab-change="loadTasks">
      <el-tab-pane label="我的待办" name="todo" />
      <el-tab-pane label="进行中" name="doing" />
      <el-tab-pane label="已完成" name="done" />
      <el-tab-pane label="已逾期" name="overdue" />
      <el-tab-pane label="我关注的" name="focused" />
    </el-tabs>

    <el-table :data="tasks" border size="small">
      <el-table-column prop="title" label="任务名称" min-width="180" />
      <el-table-column label="来源" width="110">
        <template #default="{ row }">
          <el-tag size="small" :type="sourceTag(row.sourceType)">{{ sourceLabel(row.sourceType) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="projectName" label="所属项目" width="160" />
      <el-table-column label="优先级" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="priorityTag(row.priority)">{{ priorityLabel(row.priority) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100" />
      <el-table-column prop="endDate" label="截止日期" width="120" />
      <el-table-column label="进度" width="140">
        <template #default="{ row }">
          <el-progress :percentage="row.progress" :status="row.overdue ? 'exception' : ''" />
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { getMyTasks, getFocusedTasks, getTaskStats } from '@/api/pmocker/taskCenter'

const activeTab = ref('todo')
const tasks = ref([])
const stats = reactive({ total: 0, done: 0, doneRate: 0, overdueCount: 0, todoCount: 0, doingCount: 0 })

const loadStats = async () => {
  const res = await getTaskStats()
  if (res.code === 0) Object.assign(stats, res.data)
}

const loadTasks = async () => {
  if (activeTab.value === 'focused') {
    const res = await getFocusedTasks()
    if (res.code === 0) tasks.value = res.data || []
  } else {
    const res = await getMyTasks({ status: activeTab.value })
    if (res.code === 0) tasks.value = res.data || []
  }
}

const sourceLabel = (s) => ({ project_task: '项目任务', issue_task: '问题任务', change_task: '变更任务', deliverable_task: '交付物任务' }[s] || s)
const sourceTag = (s) => ({ project_task: '', issue_task: 'warning', change_task: 'danger', deliverable_task: 'success' }[s] || '')
const priorityLabel = (p) => ({ 0: 'P0 紧急', 1: 'P1 高', 2: 'P2 中', 3: 'P3 低' }[p] || 'P2 中')
const priorityTag = (p) => ({ 0: 'danger', 1: 'warning', 2: 'info', 3: 'info' }[p] || 'info')

onMounted(() => { loadStats(); loadTasks() })
</script>

<style scoped>
.task-center { padding: 16px; }
.stat { display: flex; justify-content: space-between; align-items: center; }
.stat b { font-size: 24px; }
</style>
```

- [ ] **Step 6: 提交**

```bash
git add gva/server/service/pmocker/task_center.go gva/server/api/v1/pmocker/task_center.go gva/server/router/pmocker/business.go gva/web/src/api/pmocker/taskCenter.js gva/web/src/view/pmocker/taskcenter/index.vue
git commit -m "feat(pmocker): 个人任务中心+我关注的子视图(T15)"
```

---

## Task 5: 项目工作台（spec T16）

**Files:**
- Create: `gva/server/service/pmocker/project_workbench.go`
- Create: `gva/server/api/v1/pmocker/project_workbench.go`
- Modify: `gva/server/router/pmocker/business.go`
- Create: `gva/web/src/api/pmocker/projectWorkbench.js`
- Create: `gva/web/src/view/pmocker/workbench/index.vue`

**Interfaces:**
- Consumes: `getUserVisibilityScope`（Task 2 pmo_dashboard.go）、`PMODashboardService.GetProjectCard`（Task 2）、`PMEntity.Priority`/`OwnerID`/`CreatedBy`（Phase 1-2 Task 1）
- Produces:
  - `ProjectWorkbenchService.GetMyProjects(userID uint) (*MyProjects, error)` — 我创建的/我负责的/我参与的
  - `ProjectWorkbenchService.GetMyProjectsByStatus(userID uint, status string) ([]ProjectCard, error)` — 按状态分组
  - `ProjectWorkbenchService.GetMyFocusedProjects(userID uint) ([]ProjectCard, error)` — P0/P1，按可见性规则
  - API：`GET /pmocker/projectWorkbench/my`、`GET /pmocker/projectWorkbench/focused`

- [ ] **Step 1: 创建 project_workbench.go service**

创建 `gva/server/service/pmocker/project_workbench.go`：

```go
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

	// 我创建的：pm_entities.created_by = userID 且 entity_type=eps_node
	var created []pmocker.PMEntity
	db.Where("entity_type = ? AND created_by = ?", "eps_node", userID).Find(&created)
	for _, p := range created {
		if card, err := (&PMODashboardService{}).GetProjectCard(p.ID); err == nil {
			result.Created = append(result.Created, *card)
		}
	}

	// 我负责的：owner_id = userID 且 entity_type=eps_node
	var lead []pmocker.PMEntity
	db.Where("entity_type = ? AND owner_id = ?", "eps_node", userID).Find(&lead)
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
	q := db.Where("entity_type = ? AND priority BETWEEN ? AND ?", "eps_node", 0, 1)
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
	// 创建的 + 负责的
	db.Model(&pmocker.PMEntity{}).
		Where("entity_type = ? AND (created_by = ? OR owner_id = ?)", "eps_node", userID, userID).
		Pluck("id", &ids)
	// 团队成员参与的
	var memberProjectIDs []uint
	db.Model(&pmocker.PMEntity{}).
		Where("entity_type = ? AND owner_id = ?", "team_member", userID).
		Distinct("project_id").Pluck("project_id", &memberProjectIDs)
	ids = append(ids, memberProjectIDs...)
	return ids
}
```

- [ ] **Step 2: 创建 project_workbench API handler**

创建 `gva/server/api/v1/pmocker/project_workbench.go`：

```go
package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProjectWorkbenchApi struct{}

var projectWorkbenchService = &pmocker.ProjectWorkbenchService{}

// My 我的项目（可按 status 过滤）
// GET /pmocker/projectWorkbench/my?status=initiating|active|archived|paused
func (a *ProjectWorkbenchApi) My(c *gin.Context) {
	userID := utils.GetUserID(c)
	status := c.Query("status")
	if status != "" {
		cards, err := projectWorkbenchService.GetMyProjectsByStatus(userID, status)
		if err != nil {
			global.GVA_LOG.Error("查询项目失败", zap.Error(err))
			response.FailWithMessage("查询失败", c)
			return
		}
		response.OkWithData(cards, c)
		return
	}
	result, err := projectWorkbenchService.GetMyProjects(userID)
	if err != nil {
		global.GVA_LOG.Error("查询项目失败", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(result, c)
}

// Focused 我关注的项目（P0/P1，按可见性规则）
// GET /pmocker/projectWorkbench/focused
func (a *ProjectWorkbenchApi) Focused(c *gin.Context) {
	userID := utils.GetUserID(c)
	cards, err := projectWorkbenchService.GetMyFocusedProjects(userID)
	if err != nil {
		global.GVA_LOG.Error("查询关注项目失败", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(cards, c)
}
```

- [ ] **Step 3: 追加路由**

在 `gva/server/router/pmocker/business.go` 的 group 内追加：

```go
// 项目工作台
group.GET("projectWorkbench/my", projectWorkbenchApi.My)
group.GET("projectWorkbench/focused", projectWorkbenchApi.Focused)
```

并在 router 包内声明：

```go
var projectWorkbenchApi = pmockerApi.ProjectWorkbenchApi{}
```

- [ ] **Step 4: 编译验证**

Run: `cd gva/server && go build ./...`
Expected: 编译通过

- [ ] **Step 5: 创建前端项目工作台 API + 页面**

创建 `gva/web/src/api/pmocker/projectWorkbench.js`：

```javascript
import service from '@/utils/request'

export const getMyProjects = (params) => {
  return service({ url: '/pmocker/projectWorkbench/my', method: 'get', params })
}

export const getFocusedProjects = () => {
  return service({ url: '/pmocker/projectWorkbench/focused', method: 'get' })
}
```

创建 `gva/web/src/view/pmocker/workbench/index.vue`：

```vue
<template>
  <div class="workbench">
    <el-page-header content="项目工作台" />

    <el-tabs v-model="activeTab" style="margin-top: 12px" @tab-change="loadData">
      <el-tab-pane label="我创建的" name="created" />
      <el-tab-pane label="我负责的" name="lead" />
      <el-tab-pane label="我参与的" name="involved" />
      <el-tab-pane label="我关注的" name="focused" />
    </el-tabs>

    <el-radio-group v-model="statusFilter" size="small" style="margin-bottom: 12px" @change="loadData">
      <el-radio-button label="">全部</el-radio-button>
      <el-radio-button label="initiating">立项中</el-radio-button>
      <el-radio-button label="active">进行中</el-radio-button>
      <el-radio-button label="archived">已归档</el-radio-button>
      <el-radio-button label="paused">已暂停</el-radio-button>
    </el-radio-group>

    <el-row :gutter="12">
      <el-col v-for="card in cards" :key="card.projectId" :span="8" style="margin-bottom: 12px">
        <el-card shadow="hover" :body-style="{ padding: '16px' }">
          <div class="card-head">
            <span class="dot" :class="card.health" />
            <span class="proj-name">{{ card.projectName }}</span>
            <el-tag size="small" :type="priorityTag(card.priority)" style="margin-left: auto">{{ priorityLabel(card.priority) }}</el-tag>
          </div>
          <el-progress :percentage="Math.round(card.progress)" :color="healthColor(card.health)" style="margin: 8px 0" />
          <div class="card-row">
            <span>成本偏差：</span>
            <b :class="card.costVariance > 0 ? 'red' : 'green'">{{ card.costVariance > 0 ? '+' : '' }}{{ (card.costVariance || 0).toFixed(2) }}</b>
          </div>
          <div class="card-row">
            <span>风险数：</span><b>{{ card.riskCount }}</b>
            <span style="margin-left: 16px">负责人：</span><b>{{ card.leaderName }}</b>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getMyProjects, getFocusedProjects } from '@/api/pmocker/projectWorkbench'

const activeTab = ref('created')
const statusFilter = ref('')
const cards = ref([])

const loadData = async () => {
  if (activeTab.value === 'focused') {
    const res = await getFocusedProjects()
    if (res.code === 0) cards.value = res.data || []
  } else {
    const params = statusFilter.value ? { status: statusFilter.value } : {}
    const res = await getMyProjects(params)
    if (res.code === 0) {
      // 非过滤状态时返回 {created, lead, involved}，按 tab 取对应分类
      if (statusFilter.value) {
        cards.value = res.data || []
      } else {
        const grouped = res.data || {}
        cards.value = grouped[activeTab.value] || []
      }
    }
  }
}

const priorityLabel = (p) => ({ 0: 'P0 紧急', 1: 'P1 高', 2: 'P2 中', 3: 'P3 低' }[p] || 'P2 中')
const priorityTag = (p) => ({ 0: 'danger', 1: 'warning', 2: 'info', 3: 'info' }[p] || 'info')
const healthColor = (h) => ({ green: '#67C23A', yellow: '#E6A23C', red: '#F56C6C' }[h] || '#67C23A')

onMounted(() => { loadData() })
</script>

<style scoped>
.workbench { padding: 16px; }
.card-head { display: flex; align-items: center; }
.proj-name { font-weight: bold; margin-left: 8px; }
.dot { width: 12px; height: 12px; border-radius: 50%; display: inline-block; }
.dot.green { background: #67C23A; }
.dot.yellow { background: #E6A23C; }
.dot.red { background: #F56C6C; }
.card-row { font-size: 13px; margin: 4px 0; }
.red { color: #F56C6C; }
.green { color: #67C23A; }
</style>
```

- [ ] **Step 6: 提交**

```bash
git add gva/server/service/pmocker/project_workbench.go gva/server/api/v1/pmocker/project_workbench.go gva/server/router/pmocker/business.go gva/web/src/api/pmocker/projectWorkbench.js gva/web/src/view/pmocker/workbench/index.vue
git commit -m "feat(pmocker): 项目工作台+我关注的子视图(T16)"
```

---

## Self-Review

### 1. Spec coverage check（对照 spec 验收标准 7.1 第 13-19 项）

| Spec 验收项 | 对应 Task | 状态 |
|------------|----------|------|
| 13 项目仪表盘（进度/成本/风险/问题/资源汇总） | Task 1 GetProjectDashboard | ✅ |
| 14 PMO 看板（跨项目组合视图，健康度 RAG） | Task 2 GetPMODashboard + GetProjectHealth | ✅ |
| 15 结项归档（归档状态机，结项报告生成） | Task 3 ArchiveProject + GetCloseReport | ✅ |
| 16 个人任务中心（聚合 4 类任务，按状态分组，统计） | Task 4 GetMyTasks/GetMyTasksByStatus/GetTaskStats | ✅ |
| 17 项目工作台（我创建/负责/参与，按状态分组，项目卡片） | Task 5 GetMyProjects/GetMyProjectsByStatus | ✅ |
| 18 优先级字段（priority P0/P1/P2/P3） | Phase 1-2 Task 1 已加，本计划 Task 2/4/5 使用 | ✅ |
| 19 我关注的子视图（P0/P1，PMO_ADMIN/DEPT_LEADER/CCB_MEMBER 可见） | Task 4 GetMyFocusedTasks + Task 5 GetMyFocusedProjects + getUserVisibilityScope | ✅ |

补充覆盖：
- spec 3.7 聚合视图层 T15/T16：Task 4/5 ✅
- spec 3.8 第四层个人工作台聚合层：Task 4/5 ✅
- spec 5.0 组织架构（角色/岗位路由）：getUserVisibilityScope 基于 sys_user_authority(9001) + sys_user_positions(DEPT_LEADER/CCB_MEMBER) ✅
- 运行报告快照（spec 8 技术决策 7 混合实时+里程碑快照）：Task 1 GenerateReportSnapshot + PMReportSnapshot 表 ✅

### 2. Placeholder scan

- 无 TBD/TODO 占位 ✅
- 所有 Go/Vue 代码块包含实际实现 ✅
- 每个编译验证步骤有明确 Run/Expected ✅
- 唯一标注：Task 1 Step 4 已显式标注 `formatPeriod` 跨包调用的修正方案（apiFormatPeriod），非占位符而是明确的实现指引 ✅

### 3. Type consistency

- `ProjectCard` 在 Task 2 (pmo_dashboard.go) 定义，Task 5 (project_workbench.go) 消费 ✅
- `getUserVisibilityScope`/`VisibilityScope` 在 Task 2 定义，Task 4/5 消费 ✅
- `getAttrDecimal/getAttrInt/getAttrString/getAttrRef` 在 Task 1 (dashboard.go) 定义，Task 2/3/4 消费 ✅
- `DashboardService.calcProgressByCount/calcCostSummary/calcRiskSummary/calcResourceSummary` 在 Task 1 定义，Task 2 GetProjectHealth/GetProjectCard 消费 ✅
- `PMODashboardService.GetProjectCard` 在 Task 2 定义，Task 1 report.go + Task 5 project_workbench.go 消费 ✅
- `ReportService.GenerateReportSnapshot` 在 Task 1 定义，Task 3 ArchiveProject 消费 ✅
- `ArchiveService.GetCloseReport` 在 Task 3 定义，Task 1 report.go 的 "close" 分支消费 ✅
- `PMReportSnapshot` 在 Task 1 business.go 追加，Task 1 report.go 使用 ✅
- `PMEntity.Priority/OwnerID/CreatedBy` 字段（Phase 1-2 Task 1 产出），Task 2/4/5 一致使用 ✅
- API 路由路径：`/pmocker/dashboard/get`、`/pmocker/report/snapshot`、`/pmocker/pmo/dashboard`、`/pmocker/project/archive`、`/pmocker/project/closeReport`、`/pmocker/taskCenter/my|focused|stats`、`/pmocker/projectWorkbench/my|focused` 前端 API 与后端路由一致 ✅

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-08-02-m12-phase5-6-reports-workbench.md`.

本计划与已有计划的关系：
- Phase 1-2（T1-T10 数据骨架+计划成本资源联动）：`2026-08-02-m12-phase1-2-data-backbone.md`（已编写）
- Phase 3-4（T8-T11 基线+事件引擎）：`2026-08-02-m12-phase3-4-baseline-events.md`（待编写）
- Phase 5-6（T12-T16 报告+个人工作台）：本计划

依赖说明：本计划 Task 1 消费 Phase 1-2 产出的 `PMEntity.Priority` 字段和 `PMTimeEntry`/`PMCostActual` model；进度计算采用内置简单算法（任务数平均），未依赖 Phase 4 的 CalcProjectProgress，保证可独立编译验证。

Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?