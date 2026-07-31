package issue

import issuesvc "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_issue/service"

type ApiGroup struct {
	Api
}

var ApiGroupApp = new(ApiGroup)

var ServiceGroupApp = issuesvc.ServiceGroupApp
