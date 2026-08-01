package team

import api "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_team/api"

// teamImpl 默认 Team 实例
var teamImpl = &Team{}

type RouterGroup struct {
	Team *Team
}

var RouterGroupApp = &RouterGroup{Team: teamImpl}
var ApiGroupApp = api.ApiGroupApp
