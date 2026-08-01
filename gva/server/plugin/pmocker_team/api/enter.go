package team

import teamsvc "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_team/service"

type ApiGroup struct {
	Api
}

var ApiGroupApp = new(ApiGroup)

// ServiceGroupApp 指向 service 包的 ServiceGroupApp，供 api 层直接调用
var ServiceGroupApp = teamsvc.ServiceGroupApp
