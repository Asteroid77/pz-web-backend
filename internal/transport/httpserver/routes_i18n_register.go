package httpserver

import "github.com/gin-gonic/gin"

func (a App) registerI18nRoutes(r *gin.Engine) {
	r.GET("/api/i18n", a.handleI18n)
}
