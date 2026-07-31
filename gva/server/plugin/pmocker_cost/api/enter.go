package cost

import costService "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_cost/service"

type ApiGroup struct {
	Api
}

var ApiGroupApp = new(ApiGroup)

// ServiceGroupApp 引用 service 层的 ServiceGroupApp，供 api 方法直接调用。
var ServiceGroupApp = costService.ServiceGroupApp
