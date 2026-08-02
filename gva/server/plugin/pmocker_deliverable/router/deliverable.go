package deliverable

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type Deliverable struct{}

func (r *Deliverable) Init(public, private *gin.RouterGroup) {
	{
		g := private.Group("deliverable").Use(middleware.OperationRecord())
		g.POST("create", ApiGroupApp.Create)
		g.DELETE("delete", ApiGroupApp.Delete)
		g.PUT("update", ApiGroupApp.Update)
		g.POST("submitReview", ApiGroupApp.SubmitReview)
		g.POST("accept", ApiGroupApp.Accept)
		g.POST("reject", ApiGroupApp.Reject)
		g.POST("createVersion", ApiGroupApp.CreateVersion)
		g.POST("baseline", ApiGroupApp.Baseline)
		g.POST("checkOut", ApiGroupApp.CheckOut)
		g.POST("checkIn", ApiGroupApp.CheckIn)
	}
	{
		g := private.Group("deliverable")
		g.GET("find", ApiGroupApp.Find)
		g.GET("list", ApiGroupApp.List)
		g.GET("listVersions", ApiGroupApp.ListVersions)
		g.GET("traceReport", ApiGroupApp.TraceReport)
		g.GET("stats", ApiGroupApp.Stats)
	}
}
