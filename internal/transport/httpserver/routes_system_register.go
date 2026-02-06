package httpserver

import "github.com/gin-gonic/gin"

func (a App) registerSystemRoutes(r *gin.Engine) {
	r.GET("/api/system/check_update", a.handleCheckUpdate)
	r.POST("/api/system/perform_update", a.handlePerformUpdate)
}
