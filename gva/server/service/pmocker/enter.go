package pmocker

// ServiceGroup 聚合所有 PMocker service
type ServiceGroup struct {
	EAVService
	WorkflowService
	RBACService
	RelationService
}

// ServiceGroupApp 全局 service 入口
var ServiceGroupApp = new(ServiceGroup)
