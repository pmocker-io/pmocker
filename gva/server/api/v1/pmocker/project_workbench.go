package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProjectWorkbenchApi struct{}

// My 我的项目（可按 status 过滤）
// GET /pmocker/projectWorkbench/my?status=initiating|active|archived|paused
func (a *ProjectWorkbenchApi) My(c *gin.Context) {
	userID := utils.GetUserID(c)
	status := c.Query("status")
	if status != "" {
		cards, err := service.ProjectWorkbenchService.GetMyProjectsByStatus(userID, status)
		if err != nil {
			global.GVA_LOG.Error("查询项目失败", zap.Error(err))
			response.FailWithMessage("查询失败", c)
			return
		}
		response.OkWithData(cards, c)
		return
	}
	result, err := service.ProjectWorkbenchService.GetMyProjects(userID)
	if err != nil {
		global.GVA_LOG.Error("查询项目失败", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(result, c)
}

// Focused 我关注的项目（P0/P1，按可见性规则）
// GET /pmocker/projectWorkbench/focused
func (a *ProjectWorkbenchApi) Focused(c *gin.Context) {
	userID := utils.GetUserID(c)
	cards, err := service.ProjectWorkbenchService.GetMyFocusedProjects(userID)
	if err != nil {
		global.GVA_LOG.Error("查询关注项目失败", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(cards, c)
}
