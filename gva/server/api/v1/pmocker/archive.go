package pmocker

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ArchiveApi struct{}

// Archive 归档项目
// POST /pmocker/project/archive  body: {projectId}
func (a *ArchiveApi) Archive(c *gin.Context) {
	var req struct {
		ProjectID uint `json:"projectId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	if err := service.ArchiveService.ArchiveProject(req.ProjectID, userID); err != nil {
		global.GVA_LOG.Error("归档失败", zap.Error(err))
		response.FailWithMessage("归档失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("项目已归档", c)
}

// CloseReport 结项报告
// GET /pmocker/project/closeReport?projectId=xxx
func (a *ArchiveApi) CloseReport(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	if projectID == 0 {
		response.FailWithMessage("projectId 不能为空", c)
		return
	}
	report, err := service.ArchiveService.GetCloseReport(uint(projectID))
	if err != nil {
		global.GVA_LOG.Error("生成结项报告失败", zap.Error(err))
		response.FailWithMessage("生成结项报告失败", c)
		return
	}
	response.OkWithData(report, c)
}
