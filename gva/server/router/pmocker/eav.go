package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

// EAVRouter EAV 路由
type EAVRouter struct{}

// InitEAV 初始化 EAV 路由。
// 与其他 pmocker 业务插件的 Init(public, private) 签名保持一致：
//   - 写操作（POST/PUT/DELETE）挂 OperationRecord，记录操作日志
//   - 读操作（GET）不挂 OperationRecord
//
// 参考：aiDoc/examples/backend/router-example.md
func (r *EAVRouter) InitEAV(public *gin.RouterGroup, private *gin.RouterGroup) {
	// 写操作记录操作日志
	{
		eav := private.Group("eav").Use(middleware.OperationRecord())
		eav.POST("entity", apiGroup.EAVApi.CreateEntity)
		eav.POST("schema", apiGroup.EAVApi.RegisterSchema)
		eav.PUT("entity", apiGroup.EAVApi.UpdateEntity)
		eav.DELETE("entity/:id", apiGroup.EAVApi.DeleteEntity)
	}
	// 读操作不记录操作日志
	{
		eav := private.Group("eav")
		eav.GET("entity/:id", apiGroup.EAVApi.GetEntity)
		eav.GET("entities", apiGroup.EAVApi.ListEntities)
		eav.GET("schema/:entityType", apiGroup.EAVApi.GetSchema)
	}
}
