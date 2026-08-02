package pmocker

import "github.com/gin-gonic/gin"

type PMORouter struct{}

func (r *PMORouter) InitPMO(public *gin.RouterGroup, private *gin.RouterGroup) {
	pmo := private.Group("pmo")
	pmo.GET("dashboard", apiGroup.PMOApi.Dashboard)
}
