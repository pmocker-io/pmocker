package pmocker

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/pmocker"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CostActualApi struct{}

func costActualPuint(s string) uint {
	u, _ := strconv.ParseUint(s, 10, 32)
	return uint(u)
}

func (a *CostActualApi) CreateCostActual(c *gin.Context) {
	var e pmocker.PMCostActual
	if err := c.ShouldBindJSON(&e); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.CreateCostActual(e); err != nil {
		global.GVA_LOG.Error("创建失败", zap.Error(err))
		response.FailWithMessage("创建失败", c)
		return
	}
	response.OkWithMessage("创建成功", c)
}

func (a *CostActualApi) UpdateCostActual(c *gin.Context) {
	var e pmocker.PMCostActual
	if err := c.ShouldBindJSON(&e); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if err := service.UpdateCostActual(e); err != nil {
		response.FailWithMessage("更新失败", c)
		return
	}
	response.OkWithMessage("更新成功", c)
}

func (a *CostActualApi) DeleteCostActual(c *gin.Context) {
	id := costActualPuint(c.Query("id"))
	if err := service.DeleteCostActual(id); err != nil {
		response.FailWithMessage("删除失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

func (a *CostActualApi) FindCostActual(c *gin.Context) {
	id := costActualPuint(c.Query("id"))
	e, err := service.GetCostActual(id)
	if err != nil {
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithData(e, c)
}

func (a *CostActualApi) ListCostActuals(c *gin.Context) {
	var pageInfo request.PageInfo
	if err := c.ShouldBindQuery(&pageInfo); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	projectID := costActualPuint(c.Query("projectId"))
	costType := c.Query("costType")
	list, total, err := service.ListCostActuals(projectID, costType, pageInfo.Page, pageInfo.PageSize)
	if err != nil {
		response.FailWithMessage("查询失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{List: list, Total: total, Page: pageInfo.Page, PageSize: pageInfo.PageSize}, "成功", c)
}

func (a *CostActualApi) ConfirmCostActual(c *gin.Context) {
	id := costActualPuint(c.Query("id"))
	if err := service.ConfirmCostActual(id); err != nil {
		response.FailWithMessage("确认失败", c)
		return
	}
	response.OkWithMessage("确认成功", c)
}
