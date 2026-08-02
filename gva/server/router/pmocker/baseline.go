package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type BaselineRouter struct{}

func (r *BaselineRouter) InitBaseline(public *gin.RouterGroup, private *gin.RouterGroup) {
	{
		bl := private.Group("baseline").Use(middleware.OperationRecord())
		bl.POST("create", apiGroup.BaselineApi.Create)
	}
	{
		bl := private.Group("baseline")
		bl.GET("list", apiGroup.BaselineApi.List)
		bl.GET("compare", apiGroup.BaselineApi.Compare)
	}
}
