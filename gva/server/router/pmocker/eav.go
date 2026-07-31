package pmocker

import "github.com/gin-gonic/gin"

// EAVRouter EAV 路由
type EAVRouter struct{}

// InitEAV 初始化 EAV 路由
func (r *EAVRouter) InitEAV(Router *gin.RouterGroup) {
	eav := Router.Group("eav")
	{
		eav.POST("entity", apiGroup.EAVApi.CreateEntity)
		eav.GET("entity/:id", apiGroup.EAVApi.GetEntity)
		eav.GET("entities", apiGroup.EAVApi.ListEntities)
		eav.POST("schema", apiGroup.EAVApi.RegisterSchema)
	}
}
