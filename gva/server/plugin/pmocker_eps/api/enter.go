package eps

import epssvc "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_eps/service"

type ApiGroup struct {
	Api
}

var ApiGroupApp = new(ApiGroup)

var ServiceGroupApp = epssvc.ServiceGroupApp
