package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a App) handleI18n(c *gin.Context) {
	resp := a.I18nApp.Get(c.DefaultQuery("lang", "CN"))
	c.JSON(http.StatusOK, resp)
}
