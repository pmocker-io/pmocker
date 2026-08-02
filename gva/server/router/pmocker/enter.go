package pmocker

import api "github.com/flipped-aurora/gin-vue-admin/server/api/v1/pmocker"

// RouterGroup 聚合所有 PMocker 路由
type RouterGroup struct {
	EAVRouter
	RelationRouter
	TimeEntryRouter
	CostActualRouter
	BaselineRouter
	VarianceRouter
	ProgressRouter
}

// RouterGroupApp 全局路由入口
var RouterGroupApp = new(RouterGroup)

// apiGroup API 引用
var apiGroup = api.ApiGroupApp
