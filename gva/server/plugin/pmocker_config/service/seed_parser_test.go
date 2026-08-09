package service

import (
	"testing"
)

const testSeedYAML = `
entity_type: requirement
module: requirement
name: 需求管理
fields:
  - {key: code, label: 需求编码, data_type: string}
  - {key: priority, label: 优先级, data_type: enum, options: [P0,P1,P2,P3], default: P2}
  - {key: source, label: 来源, data_type: enum, options: [客户,市场,内部,合规]}
states:
  - {status: draft, label: 草稿, tag_type: info}
  - {status: reviewing, label: 评审中, tag_type: warning}
  - {status: published, label: 已发布, tag_type: success}
transitions:
  - {from: draft, to: reviewing, action: submit}
  - {from: reviewing, to: published, action: approve}
  - {from: reviewing, to: draft, action: reject, rollback: true}
projects:
  - project_id: 3
    entities:
      requirement:
        - {title: 排产算法, status: published, priority: P0}
        - {title: 可视化看板, status: draft, priority: P1}
`

func TestParseSeedYAML(t *testing.T) {
	seed, err := ParseSeedYAML([]byte(testSeedYAML))
	if err != nil {
		t.Fatalf("ParseSeedYAML: %v", err)
	}
	if seed.EntityType != "requirement" {
		t.Fatalf("EntityType = %s, want requirement", seed.EntityType)
	}
	if len(seed.Fields) != 3 {
		t.Fatalf("Fields count = %d, want 3", len(seed.Fields))
	}
	if seed.Fields[1].Key != "priority" || len(seed.Fields[1].Options) != 4 || seed.Fields[1].Default != "P2" {
		t.Fatalf("priority 字段解析错误: %+v", seed.Fields[1])
	}
	if len(seed.States) != 3 {
		t.Fatalf("States count = %d, want 3", len(seed.States))
	}
	if len(seed.Transitions) != 3 {
		t.Fatalf("Transitions count = %d, want 3", len(seed.Transitions))
	}
	if !seed.Transitions[2].Rollback {
		t.Fatal("reject 应标记 rollback=true")
	}
	if len(seed.Projects) != 1 {
		t.Fatalf("Projects count = %d, want 1", len(seed.Projects))
	}
	if seed.Projects[0].ProjectID != 3 {
		t.Fatalf("ProjectID = %d, want 3", seed.Projects[0].ProjectID)
	}
	reqs := seed.Projects[0].Entities["requirement"]
	if len(reqs) != 2 {
		t.Fatalf("requirement entities count = %d, want 2", len(reqs))
	}
	if reqs[0]["title"] != "排产算法" {
		t.Fatalf("entity title = %v, want 排产算法", reqs[0]["title"])
	}
}

func TestParseSeedYAMLInvalid(t *testing.T) {
	if _, err := ParseSeedYAML([]byte("not: [valid yaml")); err == nil {
		t.Fatal("非法 YAML 应返回错误")
	}
	if _, err := ParseSeedYAML([]byte("")); err == nil {
		t.Fatal("空 YAML 应返回错误")
	}
}

func TestParseSeedYAMLEPSTree(t *testing.T) {
	epsYAML := `
entity_type: eps_node
module: eps
name: 组织EPS
fields:
  - {key: type, label: 节点类型, data_type: enum, options: [group,division,program,project]}
  - {key: code, label: 编码, data_type: string}
states:
  - {status: active, label: 进行中, tag_type: success}
transitions: []
projects:
  - code: GROUP_HQ
    name: 集团总部
    type: group
    children:
      - code: PROJ_A
        name: 智能排产系统研发
        type: project
        status: active
        priority: 1
`
	seed, err := ParseSeedYAML([]byte(epsYAML))
	if err != nil {
		t.Fatalf("ParseSeedYAML(eps): %v", err)
	}
	if seed.EntityType != "eps_node" {
		t.Fatalf("EntityType = %s, want eps_node", seed.EntityType)
	}
	if len(seed.Projects) != 1 {
		t.Fatalf("Projects count = %d, want 1", len(seed.Projects))
	}
	root := seed.Projects[0]
	if root.Code != "GROUP_HQ" || root.Type != "group" {
		t.Fatalf("root = %+v", root)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children count = %d, want 1", len(root.Children))
	}
	if root.Children[0].Type != "project" || root.Children[0].Status != "active" {
		t.Fatalf("child = %+v", root.Children[0])
	}
}
