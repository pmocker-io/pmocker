package risk

import risksvc "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_risk/service"

type ApiGroup struct {
	Api
}

var ApiGroupApp = new(ApiGroup)

// ServiceGroupApp 指向 service 包的 ServiceGroupApp，供 api 层直接调用
var ServiceGroupApp = risksvc.ServiceGroupApp
