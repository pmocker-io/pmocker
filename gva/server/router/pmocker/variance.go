package pmocker

import "github.com/gin-gonic/gin"

type VarianceRouter struct{}

func (r *VarianceRouter) InitVariance(public *gin.RouterGroup, private *gin.RouterGroup) {
	vr := private.Group("variance")
	vr.GET("calc", apiGroup.VarianceApi.Calc)
	vr.GET("alerts", apiGroup.VarianceApi.Alerts)
}
