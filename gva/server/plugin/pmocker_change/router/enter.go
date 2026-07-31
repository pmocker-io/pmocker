package change

import api "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_change/api"

var changeImpl = &Change{}

type RouterGroup struct {
	Change *Change
}

var RouterGroupApp = &RouterGroup{Change: changeImpl}
var ApiGroupApp = api.ApiGroupApp
