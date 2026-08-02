package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type ArchiveRouter struct{}

func (r *ArchiveRouter) InitArchive(public *gin.RouterGroup, private *gin.RouterGroup) {
	{
		project := private.Group("project").Use(middleware.OperationRecord())
		project.POST("archive", apiGroup.ArchiveApi.Archive)
	}
	{
		project := private.Group("project")
		project.GET("closeReport", apiGroup.ArchiveApi.CloseReport)
	}
}
