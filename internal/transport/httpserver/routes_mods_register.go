package httpserver

import "github.com/gin-gonic/gin"

func (a App) registerModsRoutes(r *gin.Engine) {
	r.GET("/api/mods/lookup", a.handleModsLookup)
	r.GET("/api/mods", a.handleListLocalMods)
}
