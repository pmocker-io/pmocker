package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	teamrouter "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_team/router"
	"github.com/gin-gonic/gin"
)

func Router(engine *gin.Engine) {
	public := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("pmocker")
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("pmocker")
	private.Use(middleware.JWTAuth()).Use(middleware.MustChangePwdGuard()).Use(middleware.CasbinHandler()).Use(middleware.DataScope())
	teamrouter.RouterGroupApp.Team.Init(public, private)
}
