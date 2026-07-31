package loader

import (
	"context"
	"testing"

	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	"github.com/pmocker-io/pmocker/pkg/pmocker/workflow"
)

// stubEAV 用于测试，记录灌入的实体类型和字段
type stubEAV struct {
	entityTypes map[string]bool
	fields      []eavtypes.FieldDef
}

func (s *stubEAV) LoadEntityType(ctx context.Context, t string) (*eavtypes.EntityType, error) {
	return nil, nil
}
func (s *stubEAV) LoadFieldDefs(ctx context.Context, t string) ([]eavtypes.FieldDef, error) {
	return nil, nil
}
func (s *stubEAV) RegisterEntityType(ctx context.Context, et eavtypes.EntityType) error {
	if s.entityTypes == nil {
		s.entityTypes = map[string]bool{}
	}
	s.entityTypes[et.TypeCode] = true
	return nil
}
func (s *stubEAV) RegisterFieldDef(ctx context.Context, fd eavtypes.FieldDef) error {
	s.fields = append(s.fields, fd)
	return nil
}
func (s *stubEAV) CreateEntity(ctx context.Context, e eavtypes.Entity) (uint, error) {
	return 0, nil
}
func (s *stubEAV) GetEntity(ctx context.Context, id uint) (*eavtypes.Entity, error) {
	return nil, nil
}
func (s *stubEAV) UpdateEntity(ctx context.Context, e eavtypes.Entity) error { return nil }
func (s *stubEAV) DeleteEntity(ctx context.Context, id uint) error           { return nil }
func (s *stubEAV) ListEntities(ctx context.Context, p uint, t string, o, l int) ([]eavtypes.Entity, int64, error) {
	return nil, 0, nil
}
func (s *stubEAV) AddRelation(ctx context.Context, src, dst uint, r string) error { return nil }
func (s *stubEAV) ListRelations(ctx context.Context, id uint) ([]eavtypes.Relation, error) {
	return nil, nil
}

// stubWF 用于测试，记录灌入的工作流
type stubWF struct{ defs []workflow.Definition }

func (s *stubWF) Start(ctx context.Context, eid uint, code string) (uint, error) {
	return 0, nil
}
func (s *stubWF) Execute(ctx context.Context, iid uint, action string, uid uint) error {
	return nil
}
func (s *stubWF) GetCurrentNode(ctx context.Context, iid uint) (*workflow.Node, error) {
	return nil, nil
}
func (s *stubWF) LoadDefinition(ctx context.Context, def workflow.Definition) error {
	s.defs = append(s.defs, def)
	return nil
}

func TestLoadSchema(t *testing.T) {
	eav := &stubEAV{}
	l := &Loader{EAV: eav}
	yaml := []byte(`
entity_type: requirement
module: requirement
name: 需求
icon: document
icon_color: blue
fields:
  - field_key: priority
    field_label: 优先级
    data_type: enum
    options_json: '["高","中","低"]'
  - field_key: source
    field_label: 来源
    data_type: string
states: [draft, reviewing, approved, rejected]
workflows: [requirement_review]
`)
	if err := l.LoadSchema(context.Background(), yaml); err != nil {
		t.Fatalf("LoadSchema: %v", err)
	}
	if !eav.entityTypes["requirement"] {
		t.Errorf("entity type requirement not registered")
	}
	if len(eav.fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(eav.fields))
	}
	if eav.fields[0].FieldKey != "priority" {
		t.Errorf("first field key = %s, want priority", eav.fields[0].FieldKey)
	}
}

func TestLoadWorkflow(t *testing.T) {
	wf := &stubWF{}
	l := &Loader{Workflow: wf}
	yaml := []byte(`
code: requirement_review
name: 需求评审流
entity_type: requirement
trigger: on_create
nodes:
  - {code: start, type: start}
  - {code: review, type: approval, approvers: "pm,sm", quorum: 2}
  - {code: end, type: end}
transitions:
  - {from: start, to: review, on: submit}
  - {from: review, to: end, on: approve, status: approved}
  - {from: review, to: end, on: reject, status: rejected}
`)
	if err := l.LoadWorkflow(context.Background(), yaml); err != nil {
		t.Fatalf("LoadWorkflow: %v", err)
	}
	if len(wf.defs) != 1 {
		t.Fatalf("expected 1 def, got %d", len(wf.defs))
	}
	if wf.defs[0].Code != "requirement_review" {
		t.Errorf("code = %s", wf.defs[0].Code)
	}
	if len(wf.defs[0].Nodes) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(wf.defs[0].Nodes))
	}
}

func TestLoadMenu(t *testing.T) {
	var got []system.SysBaseMenu
	l := &Loader{MenuReg: func(menus ...system.SysBaseMenu) { got = menus }}
	yaml := []byte(`
menus:
  - path: requirement
    name: pmockerRequirement
    component: plugin/pmocker_requirement/view/index
    sort: 10
    title: 需求管理
    icon: document
  - path: requirement/list
    name: pmockerRequirementList
    parent_name: pmockerRequirement
    component: plugin/pmocker_requirement/view/list
    sort: 1
    title: 需求列表
    icon: list
`)
	if err := l.LoadMenu(yaml); err != nil {
		t.Fatalf("LoadMenu: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 menus, got %d", len(got))
	}
	if got[0].Name != "pmockerRequirement" {
		t.Errorf("first menu name = %s", got[0].Name)
	}
}
