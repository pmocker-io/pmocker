package issue

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	"github.com/gin-gonic/gin"
)

type Api struct{}

func (a *Api) Create(c *gin.Context) {
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
	id, err := ServiceGroupApp.CreateIssue(c.Request.Context(), req.ProjectID, req.Title, req.Attrs, req.CreatorID)
	if err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"id": id}, c)
}

func (a *Api) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误: id 必填", c)
		return
	}
	if err := ServiceGroupApp.DeleteIssue(c.Request.Context(), uint(id)); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *Api) Update(c *gin.Context) {
	var e eavtypes.Entity
	if err := c.ShouldBindJSON(&e); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.UpdateIssue(c.Request.Context(), e); err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *Api) Find(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误: id 必填", c)
		return
	}
	e, err := ServiceGroupApp.GetIssue(c.Request.Context(), uint(id))
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(e, c)
}

func (a *Api) List(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, total, err := ServiceGroupApp.ListIssues(c.Request.Context(), uint(projectID), offset, limit)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "查询成功", c)
}

func (a *Api) Board(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	board, err := ServiceGroupApp.IssueBoard(c.Request.Context(), uint(projectID))
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(board, c)
}

func (a *Api) Stats(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	stats, err := ServiceGroupApp.IssueStats(c.Request.Context(), uint(projectID))
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(stats, c)
}

func (a *Api) Assign(c *gin.Context) {
	var req struct {
		ID         uint `json:"id"`
		AssigneeID uint `json:"assigneeId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.AssignIssue(c.Request.Context(), req.ID, req.AssigneeID); err != nil {
		response.FailWithMessage("指派失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("指派成功", c)
}

func (a *Api) Resolve(c *gin.Context) {
	var req struct {
		ID         uint   `json:"id"`
		Resolution string `json:"resolution"`
		RootCause  string `json:"rootCause"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.ResolveIssue(c.Request.Context(), req.ID, req.Resolution, req.RootCause); err != nil {
		response.FailWithMessage("解决失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("解决成功", c)
}

func (a *Api) Close(c *gin.Context) {
	var req struct {
		ID uint `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.CloseIssue(c.Request.Context(), req.ID); err != nil {
		response.FailWithMessage("关闭失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("关闭成功", c)
}

func (a *Api) Reopen(c *gin.Context) {
	var req struct {
		ID uint `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.ReopenIssue(c.Request.Context(), req.ID); err != nil {
		response.FailWithMessage("重开失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("重开成功", c)
}
