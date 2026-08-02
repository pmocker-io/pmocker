package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type TimeEntryRouter struct{}

func (r *TimeEntryRouter) InitTimeEntry(public *gin.RouterGroup, private *gin.RouterGroup) {
	{
		timeEntry := private.Group("timeEntry").Use(middleware.OperationRecord())
		timeEntry.POST("create", apiGroup.TimeEntryApi.CreateTimeEntry)
		timeEntry.PUT("update", apiGroup.TimeEntryApi.UpdateTimeEntry)
		timeEntry.DELETE("delete", apiGroup.TimeEntryApi.DeleteTimeEntry)
		timeEntry.POST("submit", apiGroup.TimeEntryApi.SubmitTimeEntry)
		timeEntry.POST("approve", apiGroup.TimeEntryApi.ApproveTimeEntry)
		timeEntry.POST("reject", apiGroup.TimeEntryApi.RejectTimeEntry)
	}
	{
		timeEntry := private.Group("timeEntry")
		timeEntry.GET("find", apiGroup.TimeEntryApi.FindTimeEntry)
		timeEntry.GET("list", apiGroup.TimeEntryApi.ListTimeEntries)
		timeEntry.GET("utilization", apiGroup.TimeEntryApi.UtilizationTimeEntry)
	}
}
