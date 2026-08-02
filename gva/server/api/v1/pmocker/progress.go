package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

type ProgressApi struct{}

func (a *ProgressApi) Get(c *gin.Context) {
	projectID := parseUintParam(c.Query("projectId"))
	percent, err := service.ProgressService.CalcProjectProgress(projectID)
	if err != nil {
		response.FailWithMessage("计算失败: "+err.Error(), c)
		return
	}
	health, _ := service.ProgressService.CalcHealthStatus(projectID)
	algo := service.ProgressService.GetProjectAlgo(projectID)
	response.OkWithData(gin.H{
		"projectId":    projectID,
		"percent":      percent,
		"algo":         algo,
		"healthStatus": health,
	}, c)
}
