package pmocker_core

import (
	"context"
	_ "embed"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_core/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
	interfaces "github.com/flipped-aurora/gin-vue-admin/server/utils/plugin/v2"
	pmockerplugin "github.com/pmocker-io/pmocker/pkg/pmocker/plugin"
	"github.com/pmocker-io/pmocker/pkg/pmocker/plugin/loader"
	pmsvc "github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//go:embed pmocker/api.yaml
var apiBytes []byte

//go:embed pmocker/menu.yaml
var menuBytes []byte

var _ interfaces.Plugin = (*plugin)(nil)
var _ pmockerplugin.PMockerPlugin = (*plugin)(nil)
var Plugin = new(plugin)

type plugin struct{}

func init() {
	interfaces.Register(Plugin)
}

// Register 实现 gva Plugin 接口。
// gva 在启动时调用所有已注册插件的 Register，此时 GVA_DB 已初始化。
func (p *plugin) Register(group *gin.Engine) {
	registerTables()
	// 注册 EAV API 到 sys_apis 表，供 Casbin 权限校验使用（auto_init.go 会自动插入规则）
	utils.RegisterApis(parseApiYaml(apiBytes)...)
	// ★ 调用所有已注册 PMocker 业务插件的 InitPMocker 钩子，灌入 schema/seed/menu/workflow
	ctx := context.Background()
	plugins := make([]interface{}, 0, len(interfaces.Registered()))
	for _, p := range interfaces.Registered() {
		plugins = append(plugins, p)
	}
	if errs := pmockerplugin.InitAllPMockerPlugins(ctx, plugins); len(errs) > 0 {
		global.GVA_LOG.Error("PMocker plugin init errors", zap.Any("errors", errs))
	}
	// 注册 PMocker 核心路由（EAV），风格与其他 pmocker 业务插件一致
	initialize.Router(group)
}

// InitPMocker 灌入 core 菜单（仪表盘/工作台/任务中心/PMO/基线/偏差）
func (p *plugin) InitPMocker(ctx context.Context) error {
	l := loader.NewFromGva(
		&pmsvc.ServiceGroupApp.EAVService,
		&pmsvc.ServiceGroupApp.WorkflowService,
		utils.RegisterMenus, utils.RegisterApis, utils.RegisterDictionaries,
	)
	if err := l.LoadMenu(menuBytes); err != nil {
		return err
	}
	return nil
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
		pmocker.PMStateDef{},
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
		pmocker.PMTimeEntry{},
		pmocker.PMCostActual{},
		pmocker.PMApprovalRecord{},
		pmocker.PMReportSnapshot{},
	}
	global.GVA_DB.AutoMigrate(tables...)
}
