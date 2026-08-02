package schedule

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

// Schedule 进度管理路由
type Schedule struct{}

// Init 初始化进度管理路由
func (r *Schedule) Init(public, private *gin.RouterGroup) {
	{
		g := private.Group("schedule").Use(middleware.OperationRecord())
		g.POST("createTask", ApiGroupApp.CreateTask)
		g.POST("updateTask", ApiGroupApp.UpdateTask)
		g.POST("createMilestone", ApiGroupApp.CreateMilestone)
		g.POST("baseline", ApiGroupApp.Baseline)
		g.POST("transitionTask", ApiGroupApp.TransitionTask)
	}
	{
		g := private.Group("schedule")
		g.GET("listTasks", ApiGroupApp.ListTasks)
		g.GET("findTask", ApiGroupApp.FindTask)
		g.DELETE("deleteTask", ApiGroupApp.DeleteTask)
		g.GET("listMilestones", ApiGroupApp.ListMilestones)
		g.POST("cpm", ApiGroupApp.CPM)
	}
}
