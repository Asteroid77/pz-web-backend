package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (a App) handleCheckUpdate(c *gin.Context) {
	newVer, url, err := a.UpdateApp.CheckUpdate()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"current":      a.Build.Version,
		"commit_sha":   a.Build.CommitSHA,
		"build_time":   a.Build.BuildTime,
		"new_version":  newVer,
		"download_url": url,
	})
}

func (a App) handlePerformUpdate(c *gin.Context) {
	var req struct {
		Url string `json:"url"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := a.UpdateApp.PerformUpdate(req.Url); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "updating"})
}
