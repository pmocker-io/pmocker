package scope

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/gin-gonic/gin"
)

type Api struct{}

func (a *Api) CreateItem(c *gin.Context) {
	var req struct {
		ProjectID uint                   `json:"projectId"`
		Title     string                 `json:"title"`
		Attrs     map[string]interface{} `json:"attrs"`
		CreatorID uint                   `json:"creatorId"`
		ParentID  uint                   `json:"parentId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	id, err := ServiceGroupApp.CreateScopeItem(c.Request.Context(), req.ProjectID, req.Title, req.Attrs, req.CreatorID, req.ParentID)
	if err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"id": id}, c)
}

func (a *Api) ListItems(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if c.Query("page") != "" {
		offset = (page - 1) * pageSize
		limit = pageSize
	}
	list, total, err := ServiceGroupApp.ListScopeItems(c.Request.Context(), uint(projectID), offset, limit)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "查询成功", c)
}

func (a *Api) BuildWBS(c *gin.Context) {
	var req struct {
		ProjectID uint   `json:"projectId"`
		RootTitle string `json:"rootTitle"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	id, err := ServiceGroupApp.BuildWBS(c.Request.Context(), req.ProjectID, req.RootTitle)
	if err != nil {
		response.FailWithMessage("构建失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"rootId": id}, c)
}

func (a *Api) GetWBS(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	nodes, err := ServiceGroupApp.GetWBS(c.Request.Context(), uint(projectID))
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(nodes, c)
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
