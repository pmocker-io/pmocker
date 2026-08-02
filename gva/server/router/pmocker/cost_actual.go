package pmocker

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type CostActualRouter struct{}

func (r *CostActualRouter) InitCostActual(public *gin.RouterGroup, private *gin.RouterGroup) {
	{
		costActual := private.Group("costActual").Use(middleware.OperationRecord())
		costActual.POST("create", apiGroup.CostActualApi.CreateCostActual)
		costActual.PUT("update", apiGroup.CostActualApi.UpdateCostActual)
		costActual.DELETE("delete", apiGroup.CostActualApi.DeleteCostActual)
		costActual.POST("confirm", apiGroup.CostActualApi.ConfirmCostActual)
	}
	{
		costActual := private.Group("costActual")
		costActual.GET("find", apiGroup.CostActualApi.FindCostActual)
		costActual.GET("list", apiGroup.CostActualApi.ListCostActuals)
	}
}
