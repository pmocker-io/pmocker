package change

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type Change struct{}

func (r *Change) Init(public, private *gin.RouterGroup) {
	{
		g := private.Group("change").Use(middleware.OperationRecord())
		g.POST("create", ApiGroupApp.Create)
		g.DELETE("delete", ApiGroupApp.Delete)
		g.PUT("update", ApiGroupApp.Update)
		g.POST("analyze", ApiGroupApp.Analyze)
		g.POST("ccbReview", ApiGroupApp.CCBReview)
		g.POST("approve", ApiGroupApp.Approve)
		g.POST("reject", ApiGroupApp.Reject)
		g.POST("implement", ApiGroupApp.Implement)
		g.POST("verify", ApiGroupApp.Verify)
		g.POST("close", ApiGroupApp.Close)
	}
	{
		g := private.Group("change")
		g.GET("find", ApiGroupApp.Find)
		g.GET("list", ApiGroupApp.List)
		g.GET("listLogs", ApiGroupApp.ListLogs)
		g.GET("impactReport", ApiGroupApp.ImpactReport)
		g.GET("ccbStats", ApiGroupApp.CCBStats)
	}
}
