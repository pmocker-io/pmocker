package schedule

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	"github.com/gin-gonic/gin"
)

// Api 进度管理 API
type Api struct{}

func (a *Api) CreateTask(c *gin.Context) {
	var req struct {
		ProjectID uint                   `json:"projectId"`
		Title     string                 `json:"title"`
		Attrs     map[string]interface{} `json:"attrs"`
		CreatorID uint                   `json:"creatorId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	id, err := service.CreateTask(c.Request.Context(), req.ProjectID, req.Title, req.Attrs, req.CreatorID)
	if err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"id": id}, c)
}

func (a *Api) ListTasks(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, total, err := service.ListTasks(c.Request.Context(), uint(projectID), offset, limit)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "查询成功", c)
}

func (a *Api) UpdateTask(c *gin.Context) {
	var e eavtypes.Entity
	if err := c.ShouldBindJSON(&e); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := service.UpdateTask(c.Request.Context(), e); err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *Api) FindTask(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	e, err := service.GetTask(c.Request.Context(), uint(id))
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(e, c)
}

func (a *Api) DeleteTask(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if err := service.DeleteTask(c.Request.Context(), uint(id)); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *Api) CPM(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	res, err := service.ComputeCPM(c.Request.Context(), uint(projectID))
	if err != nil {
		response.FailWithMessage("计算失败: "+err.Error(), c)
		return
	}
	response.OkWithData(res, c)
}

func (a *Api) CreateMilestone(c *gin.Context) {
	var req struct {
		ProjectID uint                   `json:"projectId"`
		Title     string                 `json:"title"`
		Attrs     map[string]interface{} `json:"attrs"`
		CreatorID uint                   `json:"creatorId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	id, err := service.CreateMilestone(c.Request.Context(), req.ProjectID, req.Title, req.Attrs, req.CreatorID)
	if err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"id": id}, c)
}

func (a *Api) ListMilestones(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, total, err := service.ListMilestones(c.Request.Context(), uint(projectID), offset, limit)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "查询成功", c)
}

func (a *Api) Baseline(c *gin.Context) {
	var req struct {
		ProjectID    uint   `json:"projectId"`
		SnapshotJSON string `json:"snapshotJson"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	id, err := service.SetBaseline(c.Request.Context(), req.ProjectID, req.SnapshotJSON)
	if err != nil {
		response.FailWithMessage("基线失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"baselineId": id}, c)
}

// TransitionTask 任务状态流转（完成时后端自动触发关联交付物检入）
func (a *Api) TransitionTask(c *gin.Context) {
	var req struct {
		ID         uint   `json:"id"`
		Status     string `json:"status"`
		OperatorID uint   `json:"operatorId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if req.ID == 0 || req.Status == "" {
		response.FailWithMessage("参数错误: id 和 status 必填", c)
		return
	}
	if req.OperatorID == 0 {
		req.OperatorID = utils.GetUserID(c)
	}
	if err := service.Transition(c.Request.Context(), req.ID, req.Status, req.OperatorID); err != nil {
		response.FailWithMessage("状态流转失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("状态流转成功", c)
}
