package deliverable

import api "github.com/flipped-aurora/gin-vue-admin/server/plugin/pmocker_deliverable/api"

var deliverableImpl = &Deliverable{}

type RouterGroup struct {
	Deliverable *Deliverable
}

var RouterGroupApp = &RouterGroup{Deliverable: deliverableImpl}
var ApiGroupApp = api.ApiGroupApp
