package eps

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	eavtypes "github.com/pmocker-io/pmocker/pkg/pmocker/eav"
	"github.com/gin-gonic/gin"
)

type Api struct{}

func (a *Api) CreateNode(c *gin.Context) {
	var req struct {
		ProjectID uint                   `json:"projectId"`
		Name      string                 `json:"name"`
		Attrs     map[string]interface{} `json:"attrs"`
		CreatorID uint                   `json:"creatorId"`
		ParentID  uint                   `json:"parentId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	id, err := ServiceGroupApp.CreateEPSNode(c.Request.Context(), req.ProjectID, req.Name, req.Attrs, req.CreatorID, req.ParentID)
	if err != nil {
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"id": id}, c)
}

func (a *Api) UpdateNode(c *gin.Context) {
	var e eavtypes.Entity
	if err := c.ShouldBindJSON(&e); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.UpdateEPSNode(c.Request.Context(), e); err != nil {
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *Api) DeleteNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误: id 必填", c)
		return
	}
	if err := ServiceGroupApp.DeleteEPSNode(c.Request.Context(), uint(id)); err != nil {
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *Api) ListNodes(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if c.Query("page") != "" {
		offset = (page - 1) * pageSize
		limit = pageSize
	}
	list, total, err := ServiceGroupApp.ListEPSNodes(c.Request.Context(), uint(projectID), offset, limit)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "查询成功", c)
}

func (a *Api) AddMember(c *gin.Context) {
	var req struct {
		ProjectID uint                   `json:"projectId"`
		Attrs     map[string]interface{} `json:"attrs"`
		CreatorID uint                   `json:"creatorId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	id, err := ServiceGroupApp.AddMember(c.Request.Context(), req.ProjectID, req.Attrs, req.CreatorID)
	if err != nil {
		response.FailWithMessage("添加失败: "+err.Error(), c)
		return
	}
	response.OkWithData(gin.H{"id": id}, c)
}

func (a *Api) RemoveMember(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("id"), 10, 64)
	if err != nil {
		response.FailWithMessage("参数错误: id 必填", c)
		return
	}
	if err := ServiceGroupApp.RemoveMember(c.Request.Context(), uint(id)); err != nil {
		response.FailWithMessage("移除失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("移除成功", c)
}

func (a *Api) ListMembers(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if c.Query("page") != "" {
		offset = (page - 1) * pageSize
		limit = pageSize
	}
	list, total, err := ServiceGroupApp.ListMembers(c.Request.Context(), uint(projectID), offset, limit)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"list": list, "total": total}, "查询成功", c)
}

// GetTree 获取 EPS 树结构（用于仪表盘等页面）
// GET /pmocker/eps/tree?projectId=xxx
func (a *Api) GetTree(c *gin.Context) {
	projectID, _ := strconv.ParseUint(c.Query("projectId"), 10, 64)
	tree, err := ServiceGroupApp.BuildEPSTree(c.Request.Context(), uint(projectID))
	if err != nil {
		response.FailWithMessage("获取EPS树失败: "+err.Error(), c)
		return
	}
	response.OkWithData(tree, c)
}

// FindNode 获取 EPS 节点详情（含 attrs）
// GET /pmocker/eps/find?ID=xxx
func (a *Api) FindNode(c *gin.Context) {
	id, err := strconv.ParseUint(c.Query("ID"), 10, 64)
	if err != nil || id == 0 {
		response.FailWithMessage("参数错误: ID 必填", c)
		return
	}
	node, err := ServiceGroupApp.GetEPSNode(c.Request.Context(), uint(id))
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(node, c)
}

func (a *Api) MoveNode(c *gin.Context) {
	var req struct {
		NodeID        uint   `json:"nodeId"`
		NewParentPath string `json:"newParentPath"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	if err := ServiceGroupApp.MoveNode(c.Request.Context(), req.NodeID, req.NewParentPath); err != nil {
		response.FailWithMessage("移动失败: "+err.Error(), c)
		return
	}
	response.OkWithMessage("移动成功", c)
}
