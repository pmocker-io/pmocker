package schedule

import api "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_schedule/api"

// RouterGroup 聚合进度管理路由
type RouterGroup struct {
	Schedule *Schedule
}

// scheduleImpl 默认 Schedule 实例
var scheduleImpl = &Schedule{}

// RouterGroupApp 全局路由入口
var RouterGroupApp = &RouterGroup{Schedule: scheduleImpl}

// ApiGroupApp API 引用（导出给 router/schedule.go 使用）
var ApiGroupApp = api.ApiGroupApp
