package risk

import api "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_risk/api"

// riskImpl 默认 Risk 实例
var riskImpl = &Risk{}

type RouterGroup struct {
	Risk *Risk
}

var RouterGroupApp = &RouterGroup{Risk: riskImpl}
var ApiGroupApp = api.ApiGroupApp
