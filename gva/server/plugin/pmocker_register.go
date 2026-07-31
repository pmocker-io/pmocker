package plugin

import (
	// 空白导入 PMocker 插件，触发各自的 init() 注册到 gva 插件注册表
	// 核心插件（最先导入，负责初始化钩子和表）
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_core"

	// 九大业务插件（按 PMBOK 知识领域排序）
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_requirement"  // 需求管理
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_scope"        // 范围管理
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_schedule"     // 进度管理
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_cost"         // 成本管理
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_risk"         // 风险管理
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_issue"        // 问题管理
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_eps"          // 组织级EPS
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_deliverable"  // 交付物管理
	_ "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_change"       // 变更管理
)
