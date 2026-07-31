package requirement

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

// Requirement 路由单例
var Requirement = new(requirement)

type requirement struct{}

// Init 初始化需求管理路由
func (r *requirement) Init(public, private *gin.RouterGroup) {
	{
		g := private.Group("requirement").Use(middleware.OperationRecord())
		g.POST("create", ApiGroupApp.Create)
		g.DELETE("delete", ApiGroupApp.Delete)
		g.PUT("update", ApiGroupApp.Update)
		g.POST("submitReview", ApiGroupApp.SubmitReview)
		g.POST("approve", ApiGroupApp.Approve)
		g.POST("reject", ApiGroupApp.Reject)
	}
	{
		g := private.Group("requirement")
		g.GET("find", ApiGroupApp.Find)
		g.GET("list", ApiGroupApp.List)
		g.GET("traceMatrix", ApiGroupApp.TraceMatrix)
	}
}
