package pmocker

import "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"

// ApiGroup 聚合所有 PMocker API
type ApiGroup struct {
	EAVApi
	RelationApi
	TimeEntryApi
	CostActualApi
	BaselineApi
	VarianceApi
}

// ApiGroupApp 全局 API 入口
var ApiGroupApp = new(ApiGroup)

// service PMocker service 引用
var service = pmocker.ServiceGroupApp
