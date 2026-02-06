package httpserver

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (a App) handleModsLookup(c *gin.Context) {
	idsStr := c.Query("ids")
	if idsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids required"})
		return
	}

	targetIds := strings.Split(idsStr, ",")
	var results []ModInfo

	lookup, _ := a.ModsApp.Lookup(targetIds)
	for _, item := range lookup {
		if item.Err != nil {
			results = append(results, ModInfo{
				Name:       "Network Error / Invalid ID",
				WorkshopID: item.WorkshopID,
				ModID:      "?",
			})
			continue
		}
		for _, m := range item.Mods {
			results = append(results, m)
		}
	}

	c.JSON(http.StatusOK, results)
}

func (a App) handleListLocalMods(c *gin.Context) {
	localMods, _ := a.ModsApp.ListLocalMods()
	if localMods == nil {
		c.JSON(http.StatusOK, []ModInfo{})
		return
	}
	c.JSON(http.StatusOK, localMods)
}
