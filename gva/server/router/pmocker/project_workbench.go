package pmocker

import "github.com/gin-gonic/gin"

type ProjectWorkbenchRouter struct{}

func (r *ProjectWorkbenchRouter) InitProjectWorkbench(public *gin.RouterGroup, private *gin.RouterGroup) {
	pw := private.Group("projectWorkbench")
	pw.GET("my", apiGroup.ProjectWorkbenchApi.My)
	pw.GET("focused", apiGroup.ProjectWorkbenchApi.Focused)
}
