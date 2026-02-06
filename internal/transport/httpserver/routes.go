package httpserver

import (
	"io/fs"

	"github.com/gin-gonic/gin"
)

type Config struct {
	BaseDataDir string
	BaseGameDir string
	ServerName  string
	LogPath     string
	DevMode     bool
	Build       BuildInfo

	ContentFS fs.FS
}

func NewEngine(cfg Config) *gin.Engine {
	r := gin.Default()
	SetupStaticAndTemplates(r, cfg.ContentFS)

	app := NewApp(cfg.BaseDataDir, cfg.BaseGameDir, cfg.ServerName, cfg.LogPath, cfg.Build, cfg.DevMode)
	app.RegisterRoutes(r)
	return r
}
