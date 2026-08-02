package pmocker

import "github.com/gin-gonic/gin"

type TaskCenterRouter struct{}

func (r *TaskCenterRouter) InitTaskCenter(public *gin.RouterGroup, private *gin.RouterGroup) {
	tc := private.Group("taskCenter")
	tc.GET("my", apiGroup.TaskCenterApi.My)
	tc.GET("focused", apiGroup.TaskCenterApi.Focused)
	tc.GET("stats", apiGroup.TaskCenterApi.Stats)
}
