package httpserver

import "github.com/gin-gonic/gin"

func (a App) registerActionRoutes(r *gin.Engine) {
	r.POST("/api/action/update_restart", a.handleUpdateAndRestart)
}
