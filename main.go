package main

import (
	"embed"
	"os"

	"pz-web-backend/internal/infra/pzpaths"
	httpserver "pz-web-backend/internal/transport/httpserver"
)

var (
	Version    = "dev"                       // 默认值 (开发模式显示 dev)
	CommitSHA  = "unknown"                   // Git Commit Hash
	BuildTime  = "unknown"                   // 构建时间
	GithubRepo = "Asteroid77/pz-web-backend" // GitHub 仓库 (user/repo)
)

//go:embed template
var contentFS embed.FS

func main() {
	devMode := os.Getenv("DEV_MODE") == "true"

	base := pzpaths.ResolveBaseDirs(pzpaths.ResolveBaseDirsOptions{
		DevMode:       devMode,
		CWD:           mustGetwd(),
		EnvDataDir:    os.Getenv("PZ_DATA_DIR"),
		EnvInstallDir: os.Getenv("PZ_INSTALL_DIR"),
	})

	baseDataDir := base.DataDir
	baseGameDir := base.GameDir

	r := httpserver.NewEngine(httpserver.Config{
		BaseDataDir: baseDataDir,
		BaseGameDir: baseGameDir,
		ServerName:  os.Getenv("PZ_SERVER_NAME"),
		DevMode:     devMode,
		Build: httpserver.BuildInfo{
			Version:    Version,
			GithubRepo: GithubRepo,
			CommitSHA:  CommitSHA,
			BuildTime:  BuildTime,
		},
		ContentFS: contentFS,
	})
	r.Run(":10888")
}

func mustGetwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}
