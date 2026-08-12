package eps

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type Eps struct{}

func (r *Eps) Init(public, private *gin.RouterGroup) {
	{
		g := private.Group("eps").Use(middleware.OperationRecord())
		g.POST("createNode", ApiGroupApp.CreateNode)
		g.PUT("updateNode", ApiGroupApp.UpdateNode)
		g.DELETE("deleteNode", ApiGroupApp.DeleteNode)
		g.POST("addMember", ApiGroupApp.AddMember)
		g.DELETE("removeMember", ApiGroupApp.RemoveMember)
		g.POST("moveNode", ApiGroupApp.MoveNode)
	}
	{
		g := private.Group("eps")
		g.GET("listNodes", ApiGroupApp.ListNodes)
		g.GET("listMembers", ApiGroupApp.ListMembers)
		g.GET("tree", ApiGroupApp.GetTree)
		g.GET("find", ApiGroupApp.FindNode)
	}
}
