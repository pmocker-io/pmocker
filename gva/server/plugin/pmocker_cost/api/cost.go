package cost

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	"github.com/gin-gonic/gin"
)

type Api struct{}

func (a *Api) CreateItem(c *gin.Context) {
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
	id, err := ServiceGroupApp.CreateItem(c.Request.Context(), req.ProjectID, req.Title, req.Attrs, req.CreatorID)
	if err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"id": id}, c)
}

func (a *Api) ListItems(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	list, total, err := ServiceGroupApp.ListItems(c.Request.Context(), uint(projectID), offset, limit)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "查询成功", c)
}

func (a *Api) EVM(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	res, err := ServiceGroupApp.ComputeEVM(c.Request.Context(), uint(projectID))
	if err != nil {
		response.FailWithMessage("计算失败: "+err.Error(), c)
		return
	}
	response.OkWithData(res, c)
}

func (a *Api) UpdateItem(c *gin.Context) {
	var e eavtypes.Entity
	if err := c.ShouldBindJSON(&e); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.UpdateItem(c.Request.Context(), e); err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *Api) FindItem(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	e, err := ServiceGroupApp.GetItem(c.Request.Context(), uint(id))
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(e, c)
}

func (a *Api) DeleteItem(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)
	if err := ServiceGroupApp.DeleteItem(c.Request.Context(), uint(id)); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
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
	id, err := ServiceGroupApp.SetBaseline(c.Request.Context(), req.ProjectID, req.SnapshotJSON)
	if err != nil {
		response.FailWithMessage("基线失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"baselineId": id}, c)
}
