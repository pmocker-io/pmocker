package pmocker

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BaselineApi struct{}

func parseUintParam(s string) uint {
	v, _ := strconv.ParseUint(s, 10, 64)
	return uint(v)
}

func (a *BaselineApi) Create(c *gin.Context) {
	type createReq struct {
		ProjectID   uint   `json:"projectId" binding:"required"`
		Type        string `json:"type" binding:"required"`
		ChangeReqID *uint  `json:"changeReqId"`
	}
	var req createReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	userID := utils.GetUserID(c)
	id, err := service.BaselineService.CreateBaseline(req.ProjectID, req.Type, req.ChangeReqID, userID)
	if err != nil {
		global.GVA_LOG.Error("创建基线失败", zap.Error(err))
		response.FailWithMessage("创建基线失败: "+err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"baselineId": id}, "创建成功", c)
}

func (a *BaselineApi) List(c *gin.Context) {
	projectID := parseUintParam(c.Query("projectId"))
	baselineType := c.Query("type")
	list, err := service.BaselineService.ListBaselines(projectID, baselineType)
	if err != nil {
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}
	response.OkWithData(list, c)
}

func (a *BaselineApi) Compare(c *gin.Context) {
	baselineID := parseUintParam(c.Query("baselineId"))
	diffs, err := service.BaselineService.CompareBaseline(baselineID)
	if err != nil {
		response.FailWithMessage("对比失败: "+err.Error(), c)
		return
	}
	response.OkWithData(diffs, c)
}
