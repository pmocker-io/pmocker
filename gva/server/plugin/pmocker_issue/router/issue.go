package issue

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type Issue struct{}

func (r *Issue) Init(public, private *gin.RouterGroup) {
	{
		g := private.Group("issue").Use(middleware.OperationRecord())
		g.POST("create", ApiGroupApp.Create)
		g.DELETE("delete", ApiGroupApp.Delete)
		g.PUT("update", ApiGroupApp.Update)
		g.POST("assign", ApiGroupApp.Assign)
		g.POST("resolve", ApiGroupApp.Resolve)
		g.POST("close", ApiGroupApp.Close)
		g.POST("reopen", ApiGroupApp.Reopen)
	}
	{
		g := private.Group("issue")
		g.GET("find", ApiGroupApp.Find)
		g.GET("list", ApiGroupApp.List)
		g.GET("board", ApiGroupApp.Board)
		g.GET("stats", ApiGroupApp.Stats)
	}
}
