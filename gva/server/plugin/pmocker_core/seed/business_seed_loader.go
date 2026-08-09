package seed

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

//go:embed business_seed.yaml
var businessSeedBytes []byte

type BusinessSeedYAML struct {
	Projects []ProjectSeedYAML `yaml:"projects"`
}

type ProjectSeedYAML struct {
	Code         string            `yaml:"code"`
	Name         string            `yaml:"name"`
	Priority     int               `yaml:"priority"`
	DeptName     string            `yaml:"dept_name"`
	LeaderUser   string            `yaml:"leader_username"`
	StartDate    string            `yaml:"start_date"`
	EndDate      string            `yaml:"end_date"`
	ProgressAlgo string            `yaml:"progress_algo"`
	Health       string            `yaml:"health"`
	Status       string            `yaml:"status"`
	Team         []TeamSeed        `yaml:"team"`
	Scope        []ScopeSeed       `yaml:"scope"`
	Schedule     []TaskSeed        `yaml:"schedule"`
	Cost         []CostSeed        `yaml:"cost"`
	Requirement  []ModuleItemSeed  `yaml:"requirement"`
	Issue        []ModuleItemSeed  `yaml:"issue"`
	Risk         []RiskSeed        `yaml:"risk"`
	Change       []ChangeSeed      `yaml:"change"`
	Deliverable  []DeliverableSeed `yaml:"deliverable"`
}

type TeamSeed struct {
	Username   string  `yaml:"username"`
	Role       string  `yaml:"role"`
	HourlyRate float64 `yaml:"hourly_rate"`
	Allocation float64 `yaml:"allocation"`
}

type ScopeSeed struct {
	Code   string `yaml:"code"`
	Name   string `yaml:"name"`
	Parent string `yaml:"parent"`
}

type TaskSeed struct {
	Title          string   `yaml:"title"`
	Owner          string   `yaml:"owner"`
	StartDate      string   `yaml:"start_date"`
	EndDate        string   `yaml:"end_date"`
	EstimatedHours float64  `yaml:"estimated_hours"`
	Status         string   `yaml:"status"`
	Progress       int      `yaml:"progress"`
	Priority       int      `yaml:"priority"`
	Deps           []string `yaml:"deps"`
}

type CostSeed struct {
	Code     string  `yaml:"code"`
	Name     string  `yaml:"name"`
	CostType string  `yaml:"cost_type"`
	Amount   float64 `yaml:"amount"`
}

type ModuleItemSeed struct {
	Title       string `yaml:"title"`
	Priority    int    `yaml:"priority"`
	Status      string `yaml:"status"`
	Owner       string `yaml:"owner"`
	Description string `yaml:"description"`
	Severity    string `yaml:"severity"`
}

type RiskSeed struct {
	Title       string  `yaml:"title"`
	Probability float64 `yaml:"probability"`
	Impact      float64 `yaml:"impact"`
	Status      string  `yaml:"status"`
	Owner       string  `yaml:"owner"`
	Response    string  `yaml:"response"`
}

type ChangeSeed struct {
	Title       string `yaml:"title"`
	Type        string `yaml:"type"`
	Status      string `yaml:"status"`
	Priority    int    `yaml:"priority"`
	Owner       string `yaml:"owner"`
	Description string `yaml:"description"`
}

type DeliverableSeed struct {
	Code    string `yaml:"code"`
	Title   string `yaml:"title"`
	Owner   string `yaml:"owner"`
	Status  string `yaml:"status"`
	Priority int    `yaml:"priority"`
}

type runtimeCtx struct {
	db        *gorm.DB
	project   *pmocker.PMEntity
	leader    uint
	userIdMap map[string]uint
	userNick  map[string]string
	memberId  map[string]uint
	taskId    map[string]uint
	allProjects []*pmocker.PMEntity
}

func LoadBusinessSeed(ctx context.Context) error {
	db := global.GVA_DB.WithContext(ctx)

	data := businessSeedBytes
	var seed BusinessSeedYAML
	if err := yaml.Unmarshal(data, &seed); err != nil {
		return fmt.Errorf("parse yaml: %w", err)
	}

	rt := &runtimeCtx{db: db}
	rt.userIdMap, rt.userNick = buildUserMap(db)

	// 检查是否已有项目种子（eps_node 带 progress_algo 属性的才是项目）
	var existingProjects []pmocker.PMEntity
	db.Where("entity_type = ? AND id IN (SELECT entity_id FROM pm_attrs WHERE field_key = ?)", "eps_node", "progress_algo").Find(&existingProjects)

	if len(existingProjects) < len(seed.Projects) {
		// 首次加载：创建所有项目
		for i := range seed.Projects {
			if err := createProject(ctx, &seed.Projects[i], rt); err != nil {
				return fmt.Errorf("项目 %s: %w", seed.Projects[i].Code, err)
			}
		}
	} else {
		// 已有项目：加载到 rt.allProjects 供扩展数据使用
		rt.allProjects = make([]*pmocker.PMEntity, len(existingProjects))
		for i := range existingProjects {
			rt.allProjects[i] = &existingProjects[i]
		}
		// 设置 rt.leader 为第一个项目的负责人
		if len(existingProjects) > 0 && existingProjects[0].OwnerID != nil {
			rt.leader = *existingProjects[0].OwnerID
		}
	}

	// V2-4: 派生扩展数据（milestone/baselines/time_entries/cost_actuals/workflows/relations/reports/archive）
	// 以及团队角色/培训/绩效 + 清理 EPS 组织架构节点
	// 总是执行：确保扩展数据存在，清理组织节点
	if err := extendAllProjectsData(ctx, rt); err != nil {
		return fmt.Errorf("扩展种子数据失败: %w", err)
	}
	return nil
}

func buildUserMap(db *gorm.DB) (map[string]uint, map[string]string) {
	var users []system.SysUser
	db.Find(&users)
	idM := make(map[string]uint)
	nickM := make(map[string]string)
	for _, u := range users {
		idM[u.Username] = u.ID
		nickM[u.Username] = u.NickName
	}
	return idM, nickM
}

func createProject(_ context.Context, p *ProjectSeedYAML, rt *runtimeCtx) error {
	db := rt.db
	leader := rt.userIdMap[p.LeaderUser]

	projEntity := &pmocker.PMEntity{
		ProjectID:  0,
		EntityType: "eps_node",
		Title:      p.Name,
		Status:     p.Status,
		OwnerID:    &leader,
		Priority:   p.Priority,
		CreatedBy:  leader,
	}
	db.Create(projEntity)
	// GORM 对 int 零值使用 default tag（Priority 默认 2），需强制更新以确保 P0(0) 正确落库
	db.Model(projEntity).Update("priority", p.Priority)
	rt.project = projEntity
	rt.leader = leader
	projectID := projEntity.ID

	createAttrStr(db, projectID, "progress_algo", p.ProgressAlgo)
	createAttrStr(db, projectID, "health", p.Health)
	createAttrStr(db, projectID, "start_date", p.StartDate)
	createAttrStr(db, projectID, "end_date", p.EndDate)
	createAttrStr(db, projectID, "code", p.Code)
	createAttrStr(db, projectID, "dept_name", p.DeptName)

	rt.memberId = map[string]uint{}
	rt.taskId = map[string]uint{}

	for _, m := range p.Team {
		uid := rt.userIdMap[m.Username]
		nick := rt.userNick[m.Username]
		if nick == "" {
			nick = m.Username
		}
		member := &pmocker.PMEntity{
			ProjectID:  projectID,
			EntityType: "team_member",
			Title:      nick,
			Status:     "active",
			OwnerID:    &uid,
			Priority:   2,
			CreatedBy:  leader,
		}
		db.Create(member)
		rt.memberId[m.Username] = member.ID
		createAttrStr(db, member.ID, "role", m.Role)
		createAttrDec(db, member.ID, "hourly_rate", m.HourlyRate)
		alloc := m.Allocation
		if alloc == 0 {
			alloc = 1.0
		}
		allocPct := int(alloc * 100)
		createAttrInt(db, member.ID, "allocation_percent", allocPct)
		createAttrInt(db, member.ID, "user_id", int(uid))
	}

	scopeCodeId := map[string]uint{}
	for _, w := range p.Scope {
		var parentID *uint
		if w.Parent != "" {
			if id, ok := scopeCodeId[w.Parent]; ok {
				parentID = &id
			}
		}
		node := &pmocker.PMEntity{
			ProjectID:  projectID,
			EntityType: "scope_item",
			Title:      w.Name,
			Status:     "active",
			OwnerID:    &leader,
			ParentID:   parentID,
			Priority:   2,
			CreatedBy:  leader,
		}
		db.Create(node)
		scopeCodeId[w.Code] = node.ID
		createAttrStr(db, node.ID, "code", w.Code)
	}

	for _, t := range p.Schedule {
		// task owner_id 必须存 sys_user ID（与 task_center/service 查询逻辑对齐），
		// 而非 team_member 实体 ID；否则按用户过滤的"我关注/我的任务"返回空。
		ownerID := rt.userIdMap[t.Owner]
		task := &pmocker.PMEntity{
			ProjectID:  projectID,
			EntityType: "task",
			Title:      t.Title,
			Status:     t.Status,
			OwnerID:    &ownerID,
			Priority:   t.Priority,
			CreatedBy:  leader,
		}
		db.Create(task)
		// GORM 对 int 零值使用 default tag（Priority 默认 2），需强制更新以确保 P0(0) 正确落库
		db.Model(task).Update("priority", t.Priority)
		rt.taskId[t.Title] = task.ID
		createAttrStr(db, task.ID, "start_date", t.StartDate)
		createAttrStr(db, task.ID, "end_date", t.EndDate)
		createAttrDec(db, task.ID, "estimated_hours", t.EstimatedHours)
		createAttrInt(db, task.ID, "progress", t.Progress)
	}
	for _, t := range p.Schedule {
		dst := rt.taskId[t.Title]
		for _, depTitle := range t.Deps {
			src := rt.taskId[depTitle]
			if src == 0 || dst == 0 || src == dst {
				continue
			}
			db.Create(&pmocker.PMTaskLink{
				SrcTaskID: src, DstTaskID: dst, LinkType: "FS", Lag: 0,
			})
		}
	}

	for _, c := range p.Cost {
		item := &pmocker.PMEntity{
			ProjectID:  projectID,
			EntityType: "cost_item",
			Title:      c.Name,
			Status:     "planned",
			OwnerID:    &leader,
			Priority:   2,
			CreatedBy:  leader,
		}
		db.Create(item)
		createAttrStr(db, item.ID, "code", c.Code)
		createAttrStr(db, item.ID, "cost_type", c.CostType)
		createAttrDec(db, item.ID, "planned_value", c.Amount)
	}

	newGenericEntity := func(entityType, title, ownerName, status string, priority int) uint {
		owner := rt.userIdMap[ownerName]
		var ownerP *uint
		if owner != 0 {
			ownerP = &owner
		}
		e := &pmocker.PMEntity{
			ProjectID:  projectID,
			EntityType: entityType,
			Title:      title,
			Status:     status,
			OwnerID:    ownerP,
			Priority:   priority,
			CreatedBy:  leader,
		}
		db.Create(e)
		// 强制更新 priority，避免 GORM default tag 将 P0(0) 替换为默认值 2
		db.Model(e).Update("priority", priority)
		return e.ID
	}

	for _, r := range p.Requirement {
		eid := newGenericEntity("requirement", r.Title, r.Owner, r.Status, r.Priority)
		createAttrStr(db, eid, "description", r.Description)
		createAttrInt(db, eid, "priority", r.Priority)
	}
	for _, r := range p.Issue {
		eid := newGenericEntity("issue", r.Title, r.Owner, r.Status, r.Priority)
		createAttrStr(db, eid, "description", r.Description)
		createAttrStr(db, eid, "severity", r.Severity)
		// assignee 属性：与 task_center 的 loadAttrAssignedTasks 查询对齐
		if oid := rt.userIdMap[r.Owner]; oid > 0 {
			createAttrInt(db, eid, "assignee", int(oid))
		}
	}
	for _, r := range p.Risk {
		eid := newGenericEntity("risk", r.Title, r.Owner, r.Status, 2)
		createAttrDec(db, eid, "probability", r.Probability)
		createAttrDec(db, eid, "impact", r.Impact)
		score := r.Probability * r.Impact
		createAttrDec(db, eid, "risk_score", score)
		// 根据 score 计算 risk_level（与 risk schema 的 risk_level 字段对齐）
		level := "low"
		if score >= 0.5 {
			level = "high"
		} else if score >= 0.25 {
			level = "medium"
		}
		createAttrStr(db, eid, "risk_level", level)
		createAttrStr(db, eid, "response_plan", r.Response)
	}
	for _, r := range p.Change {
		eid := newGenericEntity("change_request", r.Title, r.Owner, r.Status, r.Priority)
		createAttrStr(db, eid, "change_type", r.Type)
		createAttrStr(db, eid, "description", r.Description)
		// assignee 属性：与 task_center 的 loadAttrAssignedTasks 查询对齐
		if oid := rt.userIdMap[r.Owner]; oid > 0 {
			createAttrInt(db, eid, "assignee", int(oid))
		}
	}
	for _, r := range p.Deliverable {
		eid := newGenericEntity("deliverable", r.Title, r.Owner, r.Status, r.Priority)
		createAttrStr(db, eid, "code", r.Code)
		createAttrStr(db, eid, "lock_status", "available")
		// reviewer 属性：与 task_center 的 loadAttrAssignedTasks 查询对齐
		if oid := rt.userIdMap[r.Owner]; oid > 0 {
			createAttrInt(db, eid, "reviewer", int(oid))
		}
	}

	// 收集项目实体指针供 extendAllProjectsData 使用
	rt.allProjects = append(rt.allProjects, projEntity)
	return nil
}

func createAttrStr(db *gorm.DB, eid uint, key, val string) {
	if val == "" {
		return
	}
	a := &pmocker.PMAttr{EntityID: eid, FieldKey: key, ValString: &val}
	db.Create(a)
}
func createAttrInt(db *gorm.DB, eid uint, key string, val int) {
	v := int64(val)
	a := &pmocker.PMAttr{EntityID: eid, FieldKey: key, ValInt: &v}
	db.Create(a)
}
func createAttrDec(db *gorm.DB, eid uint, key string, val float64) {
	a := &pmocker.PMAttr{EntityID: eid, FieldKey: key, ValDecimal: &val}
	db.Create(a)
}
