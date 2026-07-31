package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
)

// RegisterPMockerTables 注册 PMocker 的所有数据库表。
// 应在 RegisterTables() 之后调用（GVA_DB 已初始化）。
func RegisterPMockerTables() {
	if global.GVA_DB == nil {
		return
	}
	pmockerTables := []interface{}{
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
	global.GVA_DB.AutoMigrate(pmockerTables...)
}
