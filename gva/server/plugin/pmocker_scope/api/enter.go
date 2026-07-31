package scope

import (
	scopeSvc "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_scope/service"
)

type ApiGroup struct {
	Api
}

var ApiGroupApp = new(ApiGroup)

// ServiceGroupApp 桥接 service 包的 ServiceGroupApp，供 api 层调用
var ServiceGroupApp = scopeSvc.ServiceGroupApp
