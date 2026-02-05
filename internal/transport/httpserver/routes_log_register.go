package httpserver

import "github.com/gin-gonic/gin"

func (a App) registerLogRoutes(r *gin.Engine) {
	r.GET("/api/logs/stream", streamLogs)
}
