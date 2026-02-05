package httpserver

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func handleRestartPanel(c *gin.Context) {
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("[System] Exiting to trigger Supervisor restart...")
		os.Exit(0)
	}()

	c.JSON(200, gin.H{"status": "ok", "message": "Restaring..."})
}
