package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	reqrouter "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_requirement/router"
	"github.com/gin-gonic/gin"
)

// Router 注册需求管理插件路由
func Router(engine *gin.Engine) {
	public := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("pmocker")
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("pmocker")
	private.Use(middleware.JWTAuth()).Use(middleware.MustChangePwdGuard()).Use(middleware.CasbinHandler()).Use(middleware.DataScope())
	reqrouter.RouterGroupApp.Requirement.Init(public, private)
}
