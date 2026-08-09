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
		group.GET("entityTypes", api.ApiGroupApp.ConfigApi.ListEntityTypes)
		group.POST("entityType", api.ApiGroupApp.ConfigApi.CreateEntityType)
		group.POST("transition", api.ApiGroupApp.ConfigApi.Transition)
		group.POST("copy", api.ApiGroupApp.ConfigApi.CopyAsDraft)
		group.GET("stateDefs/public", api.ApiGroupApp.ConfigApi.ListStateDefsPublic)
		group.GET("seedEntities", api.ApiGroupApp.ConfigApi.ListSeedEntities)
		group.POST("export", api.ApiGroupApp.ConfigApi.Export)
	}
}
