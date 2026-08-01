package team

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type Team struct{}

func (r *Team) Init(public, private *gin.RouterGroup) {
	// 写操作记录操作日志
	{
		g := private.Group("team").Use(middleware.OperationRecord())
		// 团队成员
		g.POST("member/create", ApiGroupApp.CreateMember)
		g.DELETE("member/delete", ApiGroupApp.DeleteMember)
		g.PUT("member/update", ApiGroupApp.UpdateMember)
		g.POST("member/transition", ApiGroupApp.TransitionMember)
		// 角色定义
		g.POST("role/create", ApiGroupApp.CreateRole)
		g.DELETE("role/delete", ApiGroupApp.DeleteRole)
		g.PUT("role/update", ApiGroupApp.UpdateRole)
		g.POST("role/transition", ApiGroupApp.TransitionRole)
		// 培训记录
		g.POST("training/create", ApiGroupApp.CreateTraining)
		g.DELETE("training/delete", ApiGroupApp.DeleteTraining)
		g.PUT("training/update", ApiGroupApp.UpdateTraining)
		g.POST("training/transition", ApiGroupApp.TransitionTraining)
		// 绩效评估
		g.POST("performance/create", ApiGroupApp.CreatePerformance)
		g.DELETE("performance/delete", ApiGroupApp.DeletePerformance)
		g.PUT("performance/update", ApiGroupApp.UpdatePerformance)
		g.POST("performance/transition", ApiGroupApp.TransitionPerformance)
	}
	// 读操作不记录操作日志
	{
		g := private.Group("team")
		// 团队成员
		g.GET("member/find", ApiGroupApp.FindMember)
		g.GET("member/list", ApiGroupApp.ListMember)
		// 角色定义
		g.GET("role/find", ApiGroupApp.FindRole)
		g.GET("role/list", ApiGroupApp.ListRole)
		// 培训记录
		g.GET("training/find", ApiGroupApp.FindTraining)
		g.GET("training/list", ApiGroupApp.ListTraining)
		// 绩效评估
		g.GET("performance/find", ApiGroupApp.FindPerformance)
		g.GET("performance/list", ApiGroupApp.ListPerformance)
	}
}
