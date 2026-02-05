package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a App) handleIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "PZ Server Manager",
	})
}
