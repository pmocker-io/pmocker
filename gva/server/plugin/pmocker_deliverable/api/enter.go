package deliverable

import dlvsvc "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_deliverable/service"

type ApiGroup struct {
	Api
}

var ApiGroupApp = new(ApiGroup)

var ServiceGroupApp = dlvsvc.ServiceGroupApp
