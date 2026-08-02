package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type DashboardRouter struct{}

func (r *DashboardRouter) InitDashboard(public *gin.RouterGroup, private *gin.RouterGroup) {
	{
		report := private.Group("report").Use(middleware.OperationRecord())
		report.POST("snapshot", apiGroup.DashboardApi.Snapshot)
	}
	{
		dash := private.Group("dashboard")
		dash.GET("get", apiGroup.DashboardApi.Get)
	}
	{
		report := private.Group("report")
		report.GET("list", apiGroup.DashboardApi.ListSnapshots)
	}
}
