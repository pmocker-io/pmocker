package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	pmockerRouter "github.com/flipped-aurora/gin-vue-admin/server/router/pmocker"
	"github.com/gin-gonic/gin"
)

// Router 注册 PMocker 核心路由（EAV）。
// 与其他 pmocker 业务插件保持一致：
//   - 前缀 = RouterPrefix(/api) + "pmocker"
//   - 私有组中间件链与主系统 PrivateGroup 完全对齐
//
// 参考：aiDoc/modules/plugin-development.md#路由注册约束
func Router(engine *gin.Engine) {
	public := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("pmocker")
	private := engine.Group(global.GVA_CONFIG.System.RouterPrefix).Group("pmocker")
	private.Use(middleware.JWTAuth()).Use(middleware.MustChangePwdGuard()).Use(middleware.CasbinHandler()).Use(middleware.DataScope())
	pmockerRouter.RouterGroupApp.InitEAV(public, private)
}
