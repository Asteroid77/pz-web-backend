package httpserver

import "github.com/gin-gonic/gin"

func (a App) registerServiceRoutes(r *gin.Engine) {
	r.POST("/api/service/restart", handleRestartPanel)
}
