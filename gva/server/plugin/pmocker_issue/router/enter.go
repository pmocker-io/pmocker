package issue

import api "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_issue/api"

var issueImpl = &Issue{}

type RouterGroup struct {
	Issue *Issue
}

var RouterGroupApp = &RouterGroup{Issue: issueImpl}
var ApiGroupApp = api.ApiGroupApp
