package cost

import api "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_cost/api"

type RouterGroup struct {
	Cost *cost
}

var RouterGroupApp = &RouterGroup{Cost: Cost}

// ApiGroupApp 引用 api 层的 ApiGroupApp，供 router 注册处理函数。
var ApiGroupApp = api.ApiGroupApp
