package requirement

// ServiceGroup 聚合需求管理所有 service
type ServiceGroup struct {
	Service
}

// ServiceGroupApp 全局 service 入口
var ServiceGroupApp = new(ServiceGroup)
