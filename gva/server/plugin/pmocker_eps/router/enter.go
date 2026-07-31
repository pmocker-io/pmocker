package eps

import api "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_eps/api"

var epsImpl = &Eps{}

type RouterGroup struct {
	Eps *Eps
}

var RouterGroupApp = &RouterGroup{Eps: epsImpl}
var ApiGroupApp = api.ApiGroupApp
