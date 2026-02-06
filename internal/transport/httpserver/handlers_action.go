package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a App) handleUpdateAndRestart(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Updating and Restarting in background..."})
	go func() {
		if a.ConfigApp.Restarter == nil {
			return
		}
		_ = a.ConfigApp.Restarter.RestartPZServer()
	}()
}
