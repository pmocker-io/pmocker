package requirement

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
)

type Api struct{}

// Create
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
	id, err := ServiceGroupApp.CreateRequirement(c.Request.Context(), req.ProjectID, req.Title, req.Attrs, req.CreatorID)
	if err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"id": id}, c)
}

func (a *Api) Find(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	e, err := ServiceGroupApp.GetRequirement(c.Request.Context(), uint(id))
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
	list, total, err := ServiceGroupApp.ListRequirements(c.Request.Context(), uint(projectID), offset, limit)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "查询成功", c)
}

func (a *Api) Update(c *gin.Context) {
	var e eavtypes.Entity
	if err := c.ShouldBindJSON(&e); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.UpdateRequirement(c.Request.Context(), e); err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *Api) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if err := ServiceGroupApp.DeleteRequirement(c.Request.Context(), uint(id)); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *Api) TraceMatrix(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	rows, err := ServiceGroupApp.TraceMatrix(c.Request.Context(), uint(projectID))
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(rows, c)
}

// SubmitReview 提交需求到评审流
func (a *Api) SubmitReview(c *gin.Context) {
	var req struct {
		RequirementID uint `json:"requirementId"`
		UserID        uint `json:"userId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	instID, err := ServiceGroupApp.SubmitReview(c.Request.Context(), req.RequirementID, req.UserID)
	if err != nil {
		response.FailWithMessage("提交评审失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"instanceId": instID}, c)
}

// Approve 批准需求
func (a *Api) Approve(c *gin.Context) {
	var req struct {
		InstanceID uint `json:"instanceId"`
		UserID     uint `json:"userId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.Approve(c.Request.Context(), req.InstanceID, req.UserID); err != nil {
		response.FailWithMessage("批准失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("已批准", c)
}

// Reject 驳回需求
func (a *Api) Reject(c *gin.Context) {
	var req struct {
		InstanceID uint `json:"instanceId"`
		UserID     uint `json:"userId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.Reject(c.Request.Context(), req.InstanceID, req.UserID); err != nil {
		response.FailWithMessage("驳回失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("已驳回", c)
}
