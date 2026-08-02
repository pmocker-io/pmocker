package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TaskCenterApi struct{}

// My 我的任务（可按 status 过滤）
// GET /pmocker/taskCenter/my?status=todo|doing|done|overdue
func (a *TaskCenterApi) My(c *gin.Context) {
	userID := utils.GetUserID(c)
	status := c.Query("status")
	var tasks interface{}
	var err error
	if status != "" {
		tasks, err = service.TaskCenterService.GetMyTasksByStatus(userID, status)
	} else {
		tasks, err = service.TaskCenterService.GetMyTasks(userID)
	}
	if err != nil {
		global.GVA_LOG.Error("查询我的任务失败", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(tasks, c)
}

// Focused 我关注的任务（P0/P1，按可见性规则）
// GET /pmocker/taskCenter/focused
func (a *TaskCenterApi) Focused(c *gin.Context) {
	userID := utils.GetUserID(c)
	tasks, err := service.TaskCenterService.GetMyFocusedTasks(userID)
	if err != nil {
		global.GVA_LOG.Error("查询关注任务失败", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(tasks, c)
}

// Stats 任务统计
// GET /pmocker/taskCenter/stats
func (a *TaskCenterApi) Stats(c *gin.Context) {
	userID := utils.GetUserID(c)
	st, err := service.TaskCenterService.GetTaskStats(userID)
	if err != nil {
		global.GVA_LOG.Error("查询任务统计失败", zap.Error(err))
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(st, c)
}
