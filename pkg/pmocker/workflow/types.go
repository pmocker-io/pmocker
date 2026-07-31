// Package workflow 定义 PMocker 声明式工作流引擎的类型和接口。
package workflow

// NodeType 节点类型
type NodeType string

const (
	NodeStart    NodeType = "start"
	NodeApproval NodeType = "approval"
	NodeAuto     NodeType = "auto"
	NodeTask     NodeType = "task"
	NodeEnd      NodeType = "end"
)

// Definition 工作流定义（对应 pm_workflow_defs 表）
type Definition struct {
	Code        string       `json:"code" yaml:"code"`
	Name        string       `json:"name" yaml:"name"`
	EntityType  string       `json:"entity_type" yaml:"entity_type"`
	Trigger     string       `json:"trigger" yaml:"trigger"`
	Nodes       []Node       `json:"nodes" yaml:"nodes"`
	Transitions []Transition `json:"transitions" yaml:"transitions"`
	OnComplete  *OnComplete  `json:"on_complete,omitempty" yaml:"on_complete,omitempty"`
}

// Node 工作流节点
type Node struct {
	Code      string   `json:"code" yaml:"code"`
	Type      NodeType `json:"type" yaml:"type"`
	Assignee  string   `json:"assignee,omitempty" yaml:"assignee,omitempty"`
	Approvers string   `json:"approvers,omitempty" yaml:"approvers,omitempty"`
	SLAHours  int      `json:"sla_hours,omitempty" yaml:"sla_hours,omitempty"`
	Quorum    int      `json:"quorum,omitempty" yaml:"quorum,omitempty"`
	Handler   string   `json:"handler,omitempty" yaml:"handler,omitempty"`
	Actions   []string `json:"actions,omitempty" yaml:"actions,omitempty"`
}

// Transition 状态转移
type Transition struct {
	From   string `json:"from" yaml:"from"`
	To     string `json:"to" yaml:"to"`
	On     string `json:"on,omitempty" yaml:"on,omitempty"`
	Status string `json:"status,omitempty" yaml:"status,omitempty"`
}

// OnComplete 完成时动作
type OnComplete struct {
	Action string `json:"action" yaml:"action"`
	Sign   bool   `json:"sign" yaml:"sign"`
}

// Instance 工作流实例
type Instance struct {
	ID           uint   `json:"id"`
	EntityID     uint   `json:"entity_id"`
	WorkflowCode string `json:"workflow_code"`
	CurrentNode  string `json:"current_node"`
	Status       string `json:"status"`
	StartedAt    string `json:"started_at"`
	CompletedAt  string `json:"completed_at,omitempty"`
}
