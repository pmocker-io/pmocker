// Package plugin 定义 PMocker 插件扩展接口。
// PMocker 在 gva plugin.Plugin 接口基础上扩展，M2 阶段会补充 InitPMocker 钩子。
package plugin

// PMockerPlugin 扩展 gva plugin.Plugin 接口（server/utils/plugin/v2）。
// gva 原生插件不实现此接口也能正常运行（兼容）。
// M2 阶段会补充 InitPMocker(ctx) error 方法。
type PMockerPlugin interface {
}
