package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

type VarianceApi struct{}

func (a *VarianceApi) Calc(c *gin.Context) {
	projectID := parseUintParam(c.Query("projectId"))
	rpt, err := service.VarianceService.CalcVariance(projectID)
	if err != nil {
		response.FailWithMessage("计算失败: "+err.Error(), c)
		return
	}
	response.OkWithData(rpt, c)
}

func (a *VarianceApi) Alerts(c *gin.Context) {
	projectID := parseUintParam(c.Query("projectId"))
	alerts, err := service.VarianceService.GetAlerts(projectID)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(alerts, c)
}
