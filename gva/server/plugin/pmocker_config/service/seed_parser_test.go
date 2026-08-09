package service

import (
	"testing"
)

const testSeedYAML = `
name: 测试配置包
modules:
  requirement:
    entity_type: requirement
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
      - project_code: PROJ_A
        entities:
          requirement:
            - {title: 排产算法, status: published, priority: P0}
            - {title: 可视化看板, status: draft, priority: P1}
  schedule:
    entity_type: task
    name: 进度管理
    fields:
      - {key: progress, label: 完成度, data_type: decimal}
    states:
      - {status: planned, label: 计划中, tag_type: info}
    transitions: []
`

func TestParseSeedYAML(t *testing.T) {
	seed, err := ParseSeedYAML([]byte(testSeedYAML))
	if err != nil {
		t.Fatalf("ParseSeedYAML: %v", err)
	}
	if seed.Name != "测试配置包" {
		t.Fatalf("Name = %s, want 测试配置包", seed.Name)
	}
	if len(seed.Modules) != 2 {
		t.Fatalf("Modules count = %d, want 2", len(seed.Modules))
	}
	req := seed.Modules["requirement"]
	if req.EntityType != "requirement" {
		t.Fatalf("requirement.entityType = %s", req.EntityType)
	}
	if len(req.Fields) != 3 {
		t.Fatalf("requirement fields count = %d, want 3", len(req.Fields))
	}
	if req.Fields[1].Key != "priority" || len(req.Fields[1].Options) != 4 || req.Fields[1].Default != "P2" {
		t.Fatalf("priority 字段解析错误: %+v", req.Fields[1])
	}
	if len(req.States) != 3 {
		t.Fatalf("requirement states count = %d, want 3", len(req.States))
	}
	if len(req.Transitions) != 3 {
		t.Fatalf("requirement transitions count = %d, want 3", len(req.Transitions))
	}
	if !req.Transitions[2].Rollback {
		t.Fatal("reject 应标记 rollback=true")
	}
	if len(req.Projects) != 1 {
		t.Fatalf("requirement projects count = %d, want 1", len(req.Projects))
	}
	reqs := req.Projects[0].Entities["requirement"]
	if len(reqs) != 2 {
		t.Fatalf("requirement entities count = %d, want 2", len(reqs))
	}
	// schedule 模块
	sch := seed.Modules["schedule"]
	if sch.EntityType != "task" {
		t.Fatalf("schedule.entityType = %s, want task", sch.EntityType)
	}
}

func TestParseSeedYAMLInvalid(t *testing.T) {
	if _, err := ParseSeedYAML([]byte("not: [valid yaml")); err == nil {
		t.Fatal("非法 YAML 应返回错误")
	}
	if _, err := ParseSeedYAML([]byte("")); err == nil {
		t.Fatal("空 YAML 应返回错误")
	}
	if _, err := ParseSeedYAML([]byte("name: x\nmodules: {}\n")); err == nil {
		t.Fatal("空 modules 应返回错误")
	}
}

func TestSerializeSeedYAMLRoundTrip(t *testing.T) {
	seed, err := ParseSeedYAML([]byte(testSeedYAML))
	if err != nil {
		t.Fatal(err)
	}
	y, err := SerializeSeedYAML(seed)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	// 反序列化应一致
	seed2, err := ParseSeedYAML([]byte(y))
	if err != nil {
		t.Fatalf("Reparse: %v", err)
	}
	if len(seed2.Modules) != 2 {
		t.Fatalf("roundtrip modules = %d, want 2", len(seed2.Modules))
	}
	if seed2.Modules["requirement"].EntityType != "requirement" {
		t.Fatalf("roundtrip requirement.entityType = %s", seed2.Modules["requirement"].EntityType)
	}
	if len(seed2.Modules["requirement"].Projects) != 1 {
		t.Fatalf("roundtrip projects = %d", len(seed2.Modules["requirement"].Projects))
	}
}

func TestSerializeSeedYAMLInvalid(t *testing.T) {
	if _, err := SerializeSeedYAML(&ConfigPackageSeed{}); err == nil {
		t.Fatal("空 seed 序列化应报错")
	}
	if _, err := SerializeSeedYAML(nil); err == nil {
		t.Fatal("nil seed 序列化应报错")
	}
}

func TestParseSeedYAMLEPSTree(t *testing.T) {
	epsYAML := `
name: 组织配置
modules:
  eps:
    entity_type: eps_node
    name: 组织EPS
    fields:
      - {key: type, label: 节点类型, data_type: enum, options: [group,division,program,project]}
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
	eps := seed.Modules["eps"]
	if eps.EntityType != "eps_node" {
		t.Fatalf("eps.entityType = %s, want eps_node", eps.EntityType)
	}
	if len(eps.Projects) != 1 {
		t.Fatalf("eps projects count = %d, want 1", len(eps.Projects))
	}
	root := eps.Projects[0]
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
