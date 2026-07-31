// Package rbac 定义 PMocker 三维权限模型的类型。
package rbac

// Action 权限动作
type Action string

const (
	ActionCreate  Action = "create"
	ActionRead    Action = "read"
	ActionUpdate  Action = "update"
	ActionDelete  Action = "delete"
	ActionApprove Action = "approve"
)

// Decision 权限判定结果
type Decision struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// EntityStatePerm 实体状态权限规则
type EntityStatePerm struct {
	State     string   `json:"state"`
	Actions   []Action `json:"actions"`
	OwnerOnly bool     `json:"owner_only"`
}
