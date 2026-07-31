package workflow

import "context"

// Engine 工作流引擎接口（gva 层实现）
type Engine interface {
	// Start 启动工作流实例
	Start(ctx context.Context, entityID uint, defCode string) (uint, error)
	// Execute 执行转移（如 approve/reject）
	Execute(ctx context.Context, instanceID uint, action string, userID uint) error
	// GetCurrentNode 获取当前节点
	GetCurrentNode(ctx context.Context, instanceID uint) (*Node, error)
	// LoadDefinition 从 YAML 加载工作流定义
	LoadDefinition(ctx context.Context, def Definition) error
}
