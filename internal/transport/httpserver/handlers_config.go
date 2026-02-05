package httpserver

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"pz-web-backend/internal/application/configapp"
)

func (a App) handleGetServerConfig(c *gin.Context) {
	lang := strings.ToUpper(c.DefaultQuery("lang", "CN"))
	lang = a.I18nApp.ResolveLang(lang)
	items, err := a.ConfigApp.GetServerConfig(lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	filename := fmt.Sprintf("%s.ini", a.ConfigApp.ServerName)
	c.JSON(http.StatusOK, gin.H{"filename": filename, "lang": lang, "items": items})
}

func (a App) handleGetSandboxConfig(c *gin.Context) {
	lang := strings.ToUpper(c.DefaultQuery("lang", "CN"))
	lang = a.I18nApp.ResolveLang(lang)
	items, err := a.ConfigApp.GetSandboxConfig(lang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	filename := fmt.Sprintf("%s_SandboxVars.lua", a.ConfigApp.ServerName)
	c.JSON(http.StatusOK, gin.H{"filename": filename, "lang": lang, "items": items})
}

func (a App) handleSaveConfig(c *gin.Context) {
	name := configapp.SaveKind(c.Param("name"))
	if name != configapp.KindServer && name != configapp.KindSandbox {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid config type"})
		return
	}

	var req SaveRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := a.ConfigApp.Save(name, req.Items, req.Restart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.Restart {
		c.JSON(http.StatusOK, gin.H{"status": "saved_and_restarting", "message": "Save completed! Restarting server..."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "saved", "message": "Successfully saved!"})
}
