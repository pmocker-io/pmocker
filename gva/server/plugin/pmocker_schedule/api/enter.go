package schedule

import scheduleSvc "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_schedule/service"

// ApiGroup 聚合进度管理 API
type ApiGroup struct {
	Api
}

// ApiGroupApp 全局 API 入口
var ApiGroupApp = new(ApiGroup)

// service 进度管理 service 引用
var service = scheduleSvc.ServiceGroupApp
