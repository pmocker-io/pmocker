package pmocker

import "github.com/gin-gonic/gin"

type ProgressRouter struct{}

func (r *ProgressRouter) InitProgress(public *gin.RouterGroup, private *gin.RouterGroup) {
	pg := private.Group("progress")
	pg.GET("get", apiGroup.ProgressApi.Get)
}
