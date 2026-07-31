package scope

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

var Scope = new(scope)

type scope struct{}

func (r *scope) Init(public, private *gin.RouterGroup) {
	{
		g := private.Group("scope").Use(middleware.OperationRecord())
		g.POST("createItem", ApiGroupApp.CreateItem)
		g.POST("buildWBS", ApiGroupApp.BuildWBS)
		g.POST("baseline", ApiGroupApp.Baseline)
	}
	{
		g := private.Group("scope")
		g.GET("listItems", ApiGroupApp.ListItems)
		g.GET("getWBS", ApiGroupApp.GetWBS)
	}
}
