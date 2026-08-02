package seed

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

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
	db       *gorm.DB
	project  *pmocker.PMEntity
	leader   uint
	userIdMap map[string]uint
	userNick  map[string]string
	memberId  map[string]uint
	taskId    map[string]uint
}

func LoadBusinessSeed(ctx context.Context) error {
	db := global.GVA_DB.WithContext(ctx)

	var count int64
	db.Model(&pmocker.PMEntity{}).Where("entity_type = ?", "eps_node").Count(&count)
	if count >= 3 {
		return nil
	}

	yamlPath := filepath.Join("gva", "server", "plugin", "pmocker_core", "seed", "business_seed.yaml")
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		altPath := filepath.Join("plugin", "pmocker_core", "seed", "business_seed.yaml")
		if _, err2 := os.Stat(altPath); err2 == nil {
			yamlPath = altPath
		}
	}
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("read business_seed.yaml: %w", err)
	}
	var seed BusinessSeedYAML
	if err := yaml.Unmarshal(data, &seed); err != nil {
		return fmt.Errorf("parse yaml: %w", err)
	}

	rt := &runtimeCtx{db: db}
	rt.userIdMap, rt.userNick = buildUserMap(db)

	for i := range seed.Projects {
		if err := createProject(ctx, &seed.Projects[i], rt); err != nil {
			return fmt.Errorf("项目 %s: %w", seed.Projects[i].Code, err)
		}
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
		ownerID := rt.memberId[t.Owner]
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
		createAttrDec(db, item.ID, "planned_amount", c.Amount)
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
	}
	for _, r := range p.Risk {
		eid := newGenericEntity("risk", r.Title, r.Owner, r.Status, 2)
		createAttrDec(db, eid, "probability", r.Probability)
		createAttrDec(db, eid, "impact", r.Impact)
		score := r.Probability * r.Impact
		createAttrDec(db, eid, "score", score)
		createAttrStr(db, eid, "response", r.Response)
	}
	for _, r := range p.Change {
		eid := newGenericEntity("change_request", r.Title, r.Owner, r.Status, r.Priority)
		createAttrStr(db, eid, "change_type", r.Type)
		createAttrStr(db, eid, "description", r.Description)
	}
	for _, r := range p.Deliverable {
		eid := newGenericEntity("deliverable", r.Title, r.Owner, r.Status, r.Priority)
		createAttrStr(db, eid, "code", r.Code)
		createAttrStr(db, eid, "lock_status", "available")
	}
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
