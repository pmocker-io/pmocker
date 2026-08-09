package api

// ApiGroup 聚合初始配置管理 API
type ApiGroup struct{ ConfigApi }

// ApiGroupApp 全局 API 入口
var ApiGroupApp = new(ApiGroup)
