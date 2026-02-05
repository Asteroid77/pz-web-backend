package httpserver

import "github.com/gin-gonic/gin"

func (a App) registerIndexRoutes(r *gin.Engine) {
	r.GET("/", a.handleIndex)
}
