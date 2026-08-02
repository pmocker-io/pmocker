package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type RelationRouter struct{}

func (r *RelationRouter) InitRelation(public *gin.RouterGroup, private *gin.RouterGroup) {
	{
		relation := private.Group("relation").Use(middleware.OperationRecord())
		relation.POST("create", apiGroup.RelationApi.Create)
		relation.DELETE("delete", apiGroup.RelationApi.Delete)
	}
	{
		relation := private.Group("relation")
		relation.GET("list", apiGroup.RelationApi.List)
	}
	{
		taskLink := private.Group("taskLink").Use(middleware.OperationRecord())
		taskLink.POST("create", apiGroup.RelationApi.CreateTaskLink)
		taskLink.DELETE("delete", apiGroup.RelationApi.DeleteTaskLink)
	}
	{
		taskLink := private.Group("taskLink")
		taskLink.GET("list", apiGroup.RelationApi.ListTaskLinks)
	}
	{
		changeLog := private.Group("changeLog")
		changeLog.GET("list", apiGroup.RelationApi.ListChangeLogs)
	}
}
