package pmocker_config

import (
	"context"
	"embed"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_config/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/plugin/plugin-tool/utils"
	configService "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_config/service"
	interfaces "github.com/flipped-aurora/gin-vue-admin/server/utils/plugin/v2"
	pmockerplugin "github.com/pmocker-io/pmocker/pkg/pmocker/plugin"
	"github.com/pmocker-io/pmocker/pkg/pmocker/plugin/loader"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//go:embed pmocker/api.yaml
var apiBytes []byte

//go:embed pmocker/menu.yaml
var menuBytes []byte

//go:embed seed
var seedFS embed.FS

var _ interfaces.Plugin = (*plugin)(nil)
var _ pmockerplugin.PMockerPlugin = (*plugin)(nil)
var Plugin = new(plugin)

type plugin struct{}

func init() { interfaces.Register(Plugin) }

func (p *plugin) Register(group *gin.Engine) {
	initialize.Router(group)
}

// InitPMocker 灌入配置管理模块的菜单、API 注册与默认配置包种子
func (p *plugin) InitPMocker(ctx context.Context) error {
	l := loader.NewFromGva(nil, nil, utils.RegisterMenus, utils.RegisterApis, utils.RegisterDictionaries)
	if err := l.LoadMenu(menuBytes); err != nil {
		return err
	}
	if err := l.LoadAPI(apiBytes); err != nil {
		return err
	}
	return loadSeedPackages(ctx)
}

// loadSeedPackages 灌入 seed/ 目录下的默认配置包（幂等）
func loadSeedPackages(ctx context.Context) error {
	svc := &configService.ConfigPackageService{}
	err := fs.WalkDir(seedFS, "seed", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		data, err := fs.ReadFile(seedFS, path)
		if err != nil {
			return err
		}
		code := strings.TrimSuffix(filepath.Base(path), ".yaml")
		code = strings.TrimPrefix(code, "config_pkg_")
		if err := svc.LoadSeedPackage(ctx, code, string(data)); err != nil {
			global.GVA_LOG.Error("加载默认配置包失败", zap.String("code", code), zap.Error(err))
			return nil // 单个失败不阻断整体
		}
		return nil
	})
	return err
}
