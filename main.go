package main

import (
	"context"
	"embed"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	handler := httpserver.NewEngine(httpserver.Config{
		BaseDataDir: baseDataDir,
		BaseGameDir: baseGameDir,
		ServerName:  os.Getenv("PZ_SERVER_NAME"),
		LogPath:     os.Getenv("PZ_LOG_PATH"),
		DevMode:     devMode,
		Build: httpserver.BuildInfo{
			Version:    Version,
			GithubRepo: GithubRepo,
			CommitSHA:  CommitSHA,
			BuildTime:  BuildTime,
		},
		ContentFS: contentFS,
	})

	srv := &http.Server{
		Addr:              ":10888",
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}
}

func mustGetwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}
