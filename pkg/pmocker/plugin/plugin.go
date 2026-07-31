// Package plugin 定义 PMocker 插件扩展接口。
// PMocker 在 gva plugin.Plugin 接口基础上扩展，gva 原生插件不实现此接口也能正常运行（兼容）。
package plugin

import "context"

// PMockerPlugin 扩展 gva plugin.Plugin 接口（server/utils/plugin/v2）。
// gva 原生插件不实现此接口也能正常运行（兼容）。
type PMockerPlugin interface {
	// InitPMocker 在 PMSystem 启动时由内核调用。
	// 用于：灌入 schema/seed/menu/工作流到 EAV 元表 + gva 系统表。
	InitPMocker(ctx context.Context) error
}

// InitAllPMockerPlugins 扫描已注册插件，对实现 PMockerPlugin 的调用 InitPMocker。
// gva 原生插件（未实现 PMockerPlugin）跳过，不影响运行。
func InitAllPMockerPlugins(ctx context.Context, plugins []interface{}) []error {
	var errs []error
	for _, p := range plugins {
		if pmockerPlugin, ok := p.(PMockerPlugin); ok {
			if err := pmockerPlugin.InitPMocker(ctx); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
}
