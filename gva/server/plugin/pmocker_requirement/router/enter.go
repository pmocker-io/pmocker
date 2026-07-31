package requirement

import api "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_requirement/api"

// RouterGroup 聚合需求管理所有路由
type RouterGroup struct {
	Requirement *requirement
}

// RouterGroupApp 全局路由入口
var RouterGroupApp = &RouterGroup{Requirement: Requirement}

// ApiGroupApp 引用 api 层单例，供路由注册使用
var ApiGroupApp = api.ApiGroupApp
