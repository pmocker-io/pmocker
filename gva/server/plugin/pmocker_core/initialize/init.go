package initialize

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_core/seed"
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
	pmockerRouter.RouterGroupApp.InitRelation(public, private)
	pmockerRouter.RouterGroupApp.InitTimeEntry(public, private)
}

// SeedOrgData 组织架构种子数据调用框架（供启动流程或初始化脚本调用）。
// 目前仅作编译检查，不主动执行；实际执行时需在 DB 初始化完成后传入有效 context。
func SeedOrgData() error {
	if global.GVA_DB == nil {
		return nil
	}
	return seed.SeedOrgStructure(context.Background())
}
