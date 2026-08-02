// 此文件由 gen_pmocker_register.go 自动生成，请勿手动编辑。
// 新增 PMocker 插件后运行: go generate ./plugin/
//go:generate go run gen_pmocker_register.go

package plugin

import (
	// 空白导入 PMocker 插件，触发各自的 init() 注册到 gva 插件注册表
	// 核心插件（最先导入，负责初始化钩子和表）
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_core"

	// 业务插件（按目录名字母序自动生成）
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_change"       // 变更管理
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_cost"         // 成本管理
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_deliverable"  // 交付物管理
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_eps"          // 组织级EPS
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_issue"        // 问题管理
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_requirement"  // 需求管理
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_risk"         // 风险管理
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_schedule"     // 进度管理
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_scope"        // 范围管理
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_team"         // 团队管理
)
