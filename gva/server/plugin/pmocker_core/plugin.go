package pmocker_core

import (
	"context"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	pmockerRouter "github.com/flipped-aurora/gin-vue-admin/server/router/pmocker"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	interfaces "github.com/flipped-aurora/gin-vue-admin/server/utils/plugin/v2"
	pmockerplugin "github.com/pmocker-io/pmocker/pkg/pmocker/plugin"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var _ interfaces.Plugin = (*plugin)(nil)
var Plugin = new(plugin)

type plugin struct{}

func init() {
	interfaces.Register(Plugin)
}

// Register 实现 gva Plugin 接口。
// gva 在启动时调用所有已注册插件的 Register，此时 GVA_DB 已初始化。
func (p *plugin) Register(group *gin.Engine) {
	registerTables()
	// ★ 调用所有已注册 PMocker 业务插件的 InitPMocker 钩子，灌入 schema/seed/menu/workflow
	ctx := context.Background()
	plugins := make([]interface{}, 0, len(interfaces.Registered()))
	for _, p := range interfaces.Registered() {
		plugins = append(plugins, p)
	}
	if errs := pmockerplugin.InitAllPMockerPlugins(ctx, plugins); len(errs) > 0 {
		global.GVA_LOG.Error("PMocker plugin init errors", zap.Any("errors", errs))
	}
	// 注册 PMocker 路由（与其他 pmocker 插件路由一致，均在 /pmocker/ 下）
	pmockerRouter.RouterGroupApp.InitEAV(group.Group("pmocker"))
}

// registerTables 注册 PMocker 的所有数据库表。
func registerTables() {
	if global.GVA_DB == nil {
		return
	}
	tables := []interface{}{
		pmocker.PMEntityType{},
		pmocker.PMFieldDef{},
		pmocker.PMRelationType{},
		pmocker.PMFieldVersion{},
		pmocker.PMEntity{},
		pmocker.PMAttr{},
		pmocker.PMRelation{},
		pmocker.PMWBSNode{},
		pmocker.PMEPSTree{},
		pmocker.PMTaskLink{},
		pmocker.PMBaseline{},
		pmocker.PMChangeLog{},
		pmocker.PMDeliverableFile{},
		pmocker.PMWorkflowDef{},
		pmocker.PMWorkflowInstance{},
	}
	global.GVA_DB.AutoMigrate(tables...)
}
