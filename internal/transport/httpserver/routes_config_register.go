package httpserver

import "github.com/gin-gonic/gin"

func (a App) registerConfigRoutes(r *gin.Engine) {
	r.GET("/api/config/server", a.handleGetServerConfig)
	r.GET("/api/config/sandbox", a.handleGetSandboxConfig)
	r.POST("/api/config/:name", a.handleSaveConfig)
}
