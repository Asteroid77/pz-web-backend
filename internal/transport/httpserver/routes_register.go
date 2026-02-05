package httpserver

import "github.com/gin-gonic/gin"

func (a App) RegisterRoutes(r *gin.Engine) {
	a.registerIndexRoutes(r)
	a.registerConfigRoutes(r)
	a.registerActionRoutes(r)
	a.registerI18nRoutes(r)
	a.registerModsRoutes(r)
	a.registerSystemRoutes(r)
	a.registerServiceRoutes(r)
	a.registerLogRoutes(r)
}
