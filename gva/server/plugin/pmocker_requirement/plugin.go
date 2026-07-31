package pmocker_requirement

import (
	"context"
	"embed"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_requirement/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
	"github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
	interfaces "github.com/flipped-aurora/gin-vue-admin/server/utils/plugin/v2"
	"github.com/gin-gonic/gin"
	pmockerplugin "github.com/pmocker-io/pmocker/pkg/pmocker/plugin"
	"github.com/pmocker-io/pmocker/pkg/pmocker/plugin/loader"
)

//go:embed pmocker/manifest.yaml
var manifestBytes []byte

//go:embed pmocker/schema.yaml
var schemaBytes []byte

//go:embed pmocker/seed.yaml
var seedBytes []byte

//go:embed pmocker/menu.yaml
var menuBytes []byte

//go:embed pmocker/api.yaml
var apiBytes []byte

//go:embed pmocker/workflows
var workflowFS embed.FS

var _ interfaces.Plugin = (*plugin)(nil)
var _ pmockerplugin.PMockerPlugin = (*plugin)(nil)

// Plugin 需求管理插件单例
var Plugin = new(plugin)

type plugin struct{}

func init() {
	interfaces.Register(Plugin)
}

// Register 实现 gva Plugin 接口，注册路由
func (p *plugin) Register(group *gin.Engine) {
	initialize.Router(group)
}

// InitPMocker 灌入 schema/seed/menu/api/workflow
func (p *plugin) InitPMocker(ctx context.Context) error {
	l := loader.NewFromGva(
		&pmocker.ServiceGroupApp.EAVService,
		&pmocker.ServiceGroupApp.WorkflowService,
		utils.RegisterMenus,
		utils.RegisterApis,
		utils.RegisterDictionaries,
	)
	_ = manifestBytes // manifest 仅用于 .pmi 镜像打包时读取，运行期不灌表
	if err := l.LoadSchema(ctx, schemaBytes); err != nil {
		return err
	}
	if err := l.LoadSeed(ctx, seedBytes); err != nil {
		return err
	}
	if err := l.LoadMenu(menuBytes); err != nil {
		return err
	}
	if err := l.LoadAPI(apiBytes); err != nil {
		return err
	}
	return l.LoadWorkflowDir(ctx, workflowFS, "pmocker/workflows")
}
