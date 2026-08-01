package pmocker_team

import (
	"context"
	"embed"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_team/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
	"github.com/flipped-aurora/gin-vue-admin/server/service/pmocker"
	interfaces "github.com/flipped-aurora/gin-vue-admin/server/utils/plugin/v2"
	pmockerplugin "github.com/pmocker-io/pmocker/pkg/pmocker/plugin"
	"github.com/pmocker-io/pmocker/pkg/pmocker/plugin/loader"
	"github.com/gin-gonic/gin"
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
var Plugin = new(plugin)

type plugin struct{}

func init() { interfaces.Register(Plugin) }

func (p *plugin) Register(group *gin.Engine) { initialize.Router(group) }

func (p *plugin) InitPMocker(ctx context.Context) error {
	l := loader.NewFromGva(
		&pmocker.ServiceGroupApp.EAVService,
		&pmocker.ServiceGroupApp.WorkflowService,
		utils.RegisterMenus, utils.RegisterApis, utils.RegisterDictionaries,
	)
	_ = manifestBytes
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
