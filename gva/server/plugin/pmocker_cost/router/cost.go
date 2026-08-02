package cost

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

var Cost = new(cost)

type cost struct{}

func (r *cost) Init(public, private *gin.RouterGroup) {
	{
		g := private.Group("cost").Use(middleware.OperationRecord())
		g.POST("createItem", ApiGroupApp.CreateItem)
		g.POST("updateItem", ApiGroupApp.UpdateItem)
		g.POST("baseline", ApiGroupApp.Baseline)
	}
	{
		g := private.Group("cost")
		g.GET("listItems", ApiGroupApp.ListItems)
		g.GET("findItem", ApiGroupApp.FindItem)
		g.DELETE("deleteItem", ApiGroupApp.DeleteItem)
		g.POST("evm", ApiGroupApp.EVM)
	}
}
