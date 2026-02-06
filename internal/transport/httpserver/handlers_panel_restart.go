package httpserver

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	panelRestartDelay = 1 * time.Second
	panelRestartSleep = time.Sleep
	panelRestartExit  = os.Exit
)

func handleRestartPanel(c *gin.Context) {
	go func() {
		panelRestartSleep(panelRestartDelay)
		panelRestartExit(0)
	}()

	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Restarting..."})
}
