package scope

import api "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_scope/api"

// RouterGroup 聚合 scope 插件路由。
// Scope 字段为导出字段，供 initialize 包通过 RouterGroupApp.Scope.Init 调用。
type RouterGroup struct {
	Scope *scope
}

var RouterGroupApp = &RouterGroup{Scope: Scope}

// ApiGroupApp 桥接 api 包的 ApiGroupApp，供 router/scope.go 使用
var ApiGroupApp = api.ApiGroupApp
