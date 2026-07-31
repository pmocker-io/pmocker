package pmocker_core

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	pmockerRouter "github.com/flipped-aurora/gin-vue-admin/server/router/pmocker"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	interfaces "github.com/flipped-aurora/gin-vue-admin/server/utils/plugin/v2"
	"github.com/gin-gonic/gin"
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
	// 注册 PMocker 路由
	pmockerRouter.RouterGroupApp.InitEAV(group.Group("api").Group("pmocker"))
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
