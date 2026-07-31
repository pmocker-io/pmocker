package risk

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type Risk struct{}

func (r *Risk) Init(public, private *gin.RouterGroup) {
	// 写操作记录操作日志
	{
		g := private.Group("risk").Use(middleware.OperationRecord())
		g.POST("create", ApiGroupApp.Create)
		g.DELETE("delete", ApiGroupApp.Delete)
		g.PUT("update", ApiGroupApp.Update)
		g.POST("assess", ApiGroupApp.Assess)
	}
	// 读操作不记录操作日志
	{
		g := private.Group("risk")
		g.GET("find", ApiGroupApp.Find)
		g.GET("list", ApiGroupApp.List)
		g.GET("matrix", ApiGroupApp.Matrix)
	}
}
