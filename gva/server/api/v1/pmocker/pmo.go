package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type PMOApi struct{}

// Dashboard PMO 看板
// GET /pmocker/pmo/dashboard
func (a *PMOApi) Dashboard(c *gin.Context) {
	dash, err := service.PMODashboardService.GetPMODashboard()
	if err != nil {
		global.GVA_LOG.Error("获取PMO看板失败", zap.Error(err))
		response.FailWithMessage("获取PMO看板失败", c)
		return
	}
	response.OkWithData(dash, c)
}
