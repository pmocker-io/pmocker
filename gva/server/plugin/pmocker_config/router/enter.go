package router

// RouterGroup 聚合初始配置管理路由
type RouterGroup struct{ ConfigRouter }

// RouterGroupApp 全局路由入口
var RouterGroupApp = new(RouterGroup)
