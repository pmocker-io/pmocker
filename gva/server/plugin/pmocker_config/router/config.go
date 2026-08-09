package router

import (
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_config/api"
	"github.com/gin-gonic/gin"
)

// ConfigRouter 初始配置管理路由
type ConfigRouter struct{}

func (r *ConfigRouter) InitConfig(public, private *gin.RouterGroup) {
	group := private.Group("config")
	{
		// 配置包管理
		group.GET("packages", api.ApiGroupApp.ConfigApi.ListPackages)
		group.POST("package", api.ApiGroupApp.ConfigApi.CreatePackage)
		group.GET("package/:id", api.ApiGroupApp.ConfigApi.GetPackage)
		group.PUT("package/:id", api.ApiGroupApp.ConfigApi.UpdatePackageSeed)
		group.POST("package/:id/copy", api.ApiGroupApp.ConfigApi.CopyPackage)
		group.POST("package/:id/publish", api.ApiGroupApp.ConfigApi.PublishPackage)
		group.POST("package/:id/transition", api.ApiGroupApp.ConfigApi.TransitionPackage)
		group.DELETE("package/:id", api.ApiGroupApp.ConfigApi.DeletePackage)
		// 版本管理
		group.GET("package/:id/versions", api.ApiGroupApp.ConfigApi.ListPackageVersions)
		group.POST("package/:id/rollback", api.ApiGroupApp.ConfigApi.RollbackPackage)
		// 导出
		group.POST("export", api.ApiGroupApp.ConfigApi.Export)
		// 已发布状态流转（前端 statusTransitions 读取）
		group.GET("stateDefs/public", api.ApiGroupApp.ConfigApi.ListStateDefsPublic)
	}
}
