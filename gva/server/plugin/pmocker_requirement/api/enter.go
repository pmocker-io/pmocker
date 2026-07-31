package requirement

import requirementService "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_requirement/service"

// ApiGroup 聚合需求管理所有 api
type ApiGroup struct {
	Api
}

// ApiGroupApp 全局 api 入口
var ApiGroupApp = new(ApiGroup)

// ServiceGroupApp 引用 service 层单例，供 api 处理器调用
var ServiceGroupApp = requirementService.ServiceGroupApp
