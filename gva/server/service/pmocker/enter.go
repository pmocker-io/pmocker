package pmocker

// ServiceGroup 聚合所有 PMocker service
type ServiceGroup struct {
	EAVService
	WorkflowService
	RBACService
	RelationService
	TaskLinkService
	ChangeLogService
	TimeEntryService
	CostLinkService
	CostActualService
	BaselineService
	VarianceService
	ProgressService
}

// ServiceGroupApp 全局 service 入口
var ServiceGroupApp = new(ServiceGroup)
