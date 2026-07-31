package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	schrouter "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_schedule/router"
	"github.com/gin-gonic/gin"
)

// Router 初始化进度管理插件路由
func Router(engine *gin.Engine) {
	public := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("pmocker")
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("pmocker")
	private.Use(middleware.JWTAuth()).Use(middleware.MustChangePwdGuard()).Use(middleware.CasbinHandler()).Use(middleware.DataScope())
	schrouter.RouterGroupApp.Schedule.Init(public, private)
}
