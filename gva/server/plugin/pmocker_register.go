package plugin

import (
	// 空白导入 pmocker_core 插件，触发其 init() 注册到 gva 插件注册表
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_core"
)
