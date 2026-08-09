// migrate_seed 迁移现有 3 项目业务种子为配置包。
// 用法: go run ./cmd/migrate
// 读取 business_seed.yaml + 各模块 schema.yaml，生成配置包 seed_yaml 并写入 pmocker_config/seed/ 目录。
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ===== business_seed.yaml 结构 =====

type BusinessSeed struct {
	Projects []BusinessProject `yaml:"projects"`
}

type BusinessProject struct {
	Code       string   `yaml:"code"`
	Name       string   `yaml:"name"`
	Priority   int      `yaml:"priority"`
	DeptName   string   `yaml:"dept_name"`
	LeaderUser string   `yaml:"leader_username"`
	StartDate  string   `yaml:"start_date"`
	EndDate    string   `yaml:"end_date"`
	Progress   string   `yaml:"progress_algo"`
	Health     string   `yaml:"health"`
	Status     string   `yaml:"status"`
	Team       []map[string]interface{} `yaml:"team"`
	Scope      []map[string]interface{} `yaml:"scope"`
	Schedule   []map[string]interface{} `yaml:"schedule"`
	Cost       []map[string]interface{} `yaml:"cost"`
	Requirement []map[string]interface{} `yaml:"requirement"`
	Issue      []map[string]interface{} `yaml:"issue"`
	Risk       []map[string]interface{} `yaml:"risk"`
	Change     []map[string]interface{} `yaml:"change"`
	Deliverable []map[string]interface{} `yaml:"deliverable"`
}

// ===== 配置包 seed_yaml 结构 =====

// ConfigSeed 聚合配置包：一个配置包 = 所有模块的完整种子集合
type ConfigSeed struct {
	Name        string                `yaml:"name"`
	Description string                `yaml:"description,omitempty"`
	Modules     map[string]ModuleSeed `yaml:"modules"`
}

// ModuleSeed 单个模块种子定义
type ModuleSeed struct {
	EntityType  string           `yaml:"entity_type"`
	Name        string           `yaml:"name"`
	Fields      []FieldSeed      `yaml:"fields"`
	States      []StateSeed      `yaml:"states"`
	Transitions []TransitionSeed `yaml:"transitions"`
	Projects    []ProjectSeed    `yaml:"projects"`
}

type FieldSeed struct {
	Key      string   `yaml:"key" json:"key"`
	Label    string   `yaml:"label" json:"label"`
	DataType string   `yaml:"data_type" json:"data_type"`
	Options  []string `yaml:"options,omitempty" json:"options"`
	Default  string   `yaml:"default,omitempty" json:"default"`
}

type StateSeed struct {
	Status  string `yaml:"status" json:"status"`
	Label   string `yaml:"label" json:"label"`
	TagType string `yaml:"tag_type" json:"tag_type"`
}

type TransitionSeed struct {
	From     string `yaml:"from" json:"from"`
	To       string `yaml:"to" json:"to"`
	Action   string `yaml:"action" json:"action"`
	Rollback bool   `yaml:"rollback,omitempty" json:"rollback"`
}

type ProjectSeed struct {
	Code        string                                `yaml:"code" json:"code"`
	ProjectCode string                                `yaml:"project_code" json:"project_code"`
	Name        string                                `yaml:"name" json:"name"`
	Type        string                                `yaml:"type" json:"type"`
	Status      string                                `yaml:"status" json:"status"`
	Priority    int                                   `yaml:"priority" json:"priority"`
	Entities    map[string][]map[string]interface{}   `yaml:"entities,omitempty" json:"entities"`
}

// ===== schema.yaml 结构（loader 格式）=====

type SchemaYaml struct {
	EntityTypes []EntitySchema `yaml:"entity_types"`
	EntityType  string         `yaml:"entity_type"`
	Module      string         `yaml:"module"`
	Name        string         `yaml:"name"`
	Fields      []FieldDefYaml `yaml:"fields"`
	States      []string       `yaml:"states"`
}

type EntitySchema struct {
	EntityType string         `yaml:"entity_type"`
	Module     string         `yaml:"module"`
	Name       string         `yaml:"name"`
	Fields     []FieldDefYaml `yaml:"fields"`
	States     []string       `yaml:"states"`
}

type FieldDefYaml struct {
	FieldKey     string `yaml:"field_key"`
	FieldLabel   string `yaml:"field_label"`
	DataType     string `yaml:"data_type"`
	OptionsJSON  string `yaml:"options_json"`
	DefaultValue string `yaml:"default_value"`
}

// moduleEntityType 各模块对应的主实体类型（业务种子迁移用）
var moduleEntityType = map[string]string{
	"requirement": "requirement",
	"schedule":    "task",
	"risk":        "risk",
	"issue":       "issue",
	"change":      "change_request",
	"deliverable": "deliverable",
	"cost":        "cost_item",
	"scope":       "scope_item",
	"team":        "team_member",
}

var moduleName = map[string]string{
	"requirement": "需求管理",
	"schedule":    "进度管理",
	"risk":        "风险管理",
	"issue":       "问题管理",
	"change":      "变更管理",
	"deliverable": "交付物管理",
	"cost":        "成本管理",
	"scope":       "范围管理",
	"team":        "团队管理",
	"eps":         "组织EPS",
}

var stateLabel = map[string]string{
	"draft": "草稿", "reviewing": "评审中", "published": "已发布", "archived": "已归档",
	"open": "待处理", "assigned": "已指派", "in_progress": "处理中", "resolved": "已解决",
	"closed": "已关闭", "reopened": "已重开", "verified": "已验证",
	"planned": "计划中", "completed": "已完成", "overdue": "已逾期", "cancelled": "已取消", "on_hold": "暂停",
	"active": "进行中", "identified": "已识别", "responding": "应对中",
	"submitted": "已提交", "approved": "已批准", "rejected": "已驳回", "implementing": "实施中",
	"accepted": "已接收", "fulfilled": "已履行", "done": "已完成",
	"proposed": "已提议", "ccb_review": "CCB评审",
}

func stateLabelFor(s string) string {
	if l, ok := stateLabel[s]; ok {
		return l
	}
	return s
}

func main() {
	root := os.Getenv("PMOCKER_SERVER_DIR")
	if root == "" {
		// 默认相对 gva/server
		wd, _ := os.Getwd()
		root = filepath.Join(wd, "..", "..")
	}
	businessYAML := filepath.Join(root, "plugin", "pmocker_core", "seed", "business_seed.yaml")
	buf, err := os.ReadFile(businessYAML)
	if err != nil {
		fmt.Println("读取 business_seed.yaml 失败:", err)
		os.Exit(1)
	}
	// 去 BOM
	buf = []byte(strings.TrimPrefix(string(buf), "\ufeff"))
	var seed BusinessSeed
	if err := yaml.Unmarshal(buf, &seed); err != nil {
		fmt.Println("解析 business_seed.yaml 失败:", err)
		os.Exit(1)
	}
	fmt.Printf("读取 %d 个项目\n", len(seed.Projects))

	// 生成单个聚合配置包（所有模块的完整种子集合）
	agg := buildAggregateConfig(seed.Projects, root)
	b, err := yaml.Marshal(agg)
	if err != nil {
		fmt.Println("marshal 失败:", err)
		os.Exit(1)
	}
	writePackage("pmbok6-hybrid", b)
	fmt.Println("聚合配置包 seed_yaml 已写入 gva/server/plugin/pmocker_config/seed/")
}

// buildAggregateConfig 生成聚合配置包：一个配置包 = 所有模块的完整种子
func buildAggregateConfig(projects []BusinessProject, root string) *ConfigSeed {
	agg := &ConfigSeed{
		Name:        "PMBOK 第六版混合型配置",
		Description: "PMocker 默认配置包：包含全部 10 个模块的实体类型/字段/状态/流转/项目种子",
		Modules:     map[string]ModuleSeed{},
	}

	// 1. EPS 模块（树层级）
	epsProjects := []ProjectSeed{}
	for _, p := range projects {
		epsProjects = append(epsProjects, ProjectSeed{
			Code: p.Code, Name: p.Name, Type: "project",
			Status: p.Status, Priority: p.Priority,
		})
	}
	agg.Modules["eps"] = ModuleSeed{
		EntityType: "eps_node",
		Name:       "组织EPS",
		Fields: []FieldSeed{
			{Key: "type", Label: "节点类型", DataType: "enum", Options: []string{"group", "division", "program", "project", "subproject"}},
			{Key: "code", Label: "编码", DataType: "string"},
			{Key: "progress_algo", Label: "完成度算法", DataType: "enum", Options: []string{"hours", "wbs", "count"}},
			{Key: "health", Label: "健康度", DataType: "enum", Options: []string{"green", "yellow", "red"}},
		},
		States: []StateSeed{
			{Status: "active", Label: "进行中", TagType: "success"},
			{Status: "archived", Label: "已归档", TagType: "info"},
			{Status: "paused", Label: "已暂停", TagType: "warning"},
		},
		Projects: epsProjects,
	}

	// 2. 各业务模块
	for module, et := range moduleEntityType {
		schema, err := loadSchema(root, module, et)
		if err != nil {
			fmt.Printf("跳过模块 %s: %v\n", module, err)
			continue
		}
		ms := ModuleSeed{
			EntityType:  et,
			Name:        moduleName[module],
			Fields:      []FieldSeed{},
			States:      []StateSeed{},
			Transitions: []TransitionSeed{},
			Projects:    []ProjectSeed{},
		}
		// 字段
		for _, f := range schema.Fields {
			fs := FieldSeed{Key: f.FieldKey, Label: f.FieldLabel, DataType: f.DataType, Default: f.DefaultValue}
			if f.OptionsJSON != "" {
				var opts []string
				if err := json.Unmarshal([]byte(f.OptionsJSON), &opts); err == nil {
					fs.Options = opts
				}
			}
			ms.Fields = append(ms.Fields, fs)
		}
		// 状态
		for _, s := range schema.States {
			ms.States = append(ms.States, StateSeed{Status: s, Label: stateLabelFor(s), TagType: stateTag(s)})
		}
		// 简化流转
		ms.Transitions = buildDefaultTransitions(schema.States)
		// 项目实体种子
		for _, p := range projects {
			ps := ProjectSeed{ProjectCode: p.Code, Name: p.Name, Type: "project", Status: p.Status, Priority: p.Priority, Entities: map[string][]map[string]interface{}{}}
			ps.Entities[et] = extractEntities(module, p)
			ms.Projects = append(ms.Projects, ps)
		}
		agg.Modules[module] = ms
	}
	return agg
}

// loadSchema 读取模块 schema.yaml 的字段/状态
func loadSchema(root, module, entityType string) (*SchemaYaml, error) {
	// 模块目录名与 module 一致（pmocker_<module>）
	modDir := "pmocker_" + module
	path := filepath.Join(root, "plugin", modDir, "pmocker", "schema.yaml")
	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var schema SchemaYaml
	if err := yaml.Unmarshal(buf, &schema); err != nil {
		return nil, err
	}
	// 多实体模式：找匹配 entity_type
	for _, es := range schema.EntityTypes {
		if es.EntityType == entityType {
			return &SchemaYaml{EntityType: es.EntityType, Module: es.Module, Name: es.Name, Fields: es.Fields, States: es.States}, nil
		}
	}
	// 单实体模式
	if schema.EntityType == entityType {
		return &schema, nil
	}
	return nil, fmt.Errorf("schema.yaml 中无 entity_type=%s", entityType)
}

// extractEntities 从项目提取该模块的实体种子
func extractEntities(module string, p BusinessProject) []map[string]interface{} {
	switch module {
	case "requirement":
		return p.Requirement
	case "schedule":
		return p.Schedule
	case "risk":
		return p.Risk
	case "issue":
		return p.Issue
	case "change":
		return p.Change
	case "deliverable":
		return p.Deliverable
	case "cost":
		return p.Cost
	case "scope":
		return p.Scope
	case "team":
		return p.Team
	}
	return nil
}

// buildDefaultTransitions 生成简化流转
func buildDefaultTransitions(states []string) []TransitionSeed {
	has := func(s string) bool {
		for _, x := range states {
			if x == s {
				return true
			}
		}
		return false
	}
	var ts []TransitionSeed
	if has("draft") {
		if has("reviewing") {
			ts = append(ts, TransitionSeed{From: "draft", To: "reviewing", Action: "submit"})
		} else if has("submitted") {
			ts = append(ts, TransitionSeed{From: "draft", To: "submitted", Action: "submit"})
		}
	}
	if has("reviewing") {
		ts = append(ts, TransitionSeed{From: "reviewing", To: "published", Action: "approve"})
		ts = append(ts, TransitionSeed{From: "reviewing", To: "draft", Action: "reject", Rollback: true})
	}
	if has("submitted") && has("approved") {
		ts = append(ts, TransitionSeed{From: "submitted", To: "approved", Action: "approve"})
		ts = append(ts, TransitionSeed{From: "submitted", To: "draft", Action: "reject", Rollback: true})
	}
	return ts
}

func stateTag(s string) string {
	switch s {
	case "completed", "closed", "approved", "published", "resolved", "verified", "accepted", "done", "fulfilled":
		return "success"
	case "in_progress", "reviewing", "assigned", "responding", "implementing", "ccb_review", "submitted":
		return "warning"
	case "rejected", "overdue", "reopened":
		return "danger"
	default:
		return "info"
	}
}

// writePackage 写配置包 seed_yaml 到 seed/ 目录
func writePackage(module string, seedYAML []byte) {
	wd, _ := os.Getwd()
	// cwd 为 gva/server（通过 PMOCKER_SERVER_DIR 或相对），seed 目录 = gva/server/plugin/pmocker_config/seed
	outDir := filepath.Join(wd, "plugin", "pmocker_config", "seed")
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Println("创建 seed 目录失败:", err)
		return
	}
	fname := filepath.Join(outDir, "config_pkg_"+module+".yaml")
	if err := os.WriteFile(fname, seedYAML, 0644); err != nil {
		fmt.Println("写入失败:", err)
		return
	}
	fmt.Printf("  %s -> %s\n", module, fname)
}
