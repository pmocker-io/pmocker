package pmocker_config

import (
	"context"
	_ "embed"

	"github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_config/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
	interfaces "github.com/flipped-aurora/gin-vue-admin/server/utils/plugin/v2"
	pmockerplugin "github.com/pmocker-io/pmocker/pkg/pmocker/plugin"
	"github.com/pmocker-io/pmocker/pkg/pmocker/plugin/loader"
	"github.com/gin-gonic/gin"
)

//go:embed pmocker/api.yaml
var apiBytes []byte

//go:embed pmocker/menu.yaml
var menuBytes []byte

var _ interfaces.Plugin = (*plugin)(nil)
var _ pmockerplugin.PMockerPlugin = (*plugin)(nil)
var Plugin = new(plugin)

type plugin struct{}

func init() { interfaces.Register(Plugin) }

func (p *plugin) Register(group *gin.Engine) {
	initialize.Router(group)
}

// InitPMocker 灌入配置管理模块的菜单与 API 注册
func (p *plugin) InitPMocker(ctx context.Context) error {
	l := loader.NewFromGva(nil, nil, utils.RegisterMenus, utils.RegisterApis, utils.RegisterDictionaries)
	if err := l.LoadMenu(menuBytes); err != nil {
		return err
	}
	return l.LoadAPI(apiBytes)
}
