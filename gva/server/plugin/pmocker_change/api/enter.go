package change

import changesvc "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_change/service"

type ApiGroup struct {
	Api
}

var ApiGroupApp = new(ApiGroup)

var ServiceGroupApp = changesvc.ServiceGroupApp
