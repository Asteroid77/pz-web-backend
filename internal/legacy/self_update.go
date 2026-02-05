package legacy

import (
	"net/http"
	"sync"
	"time"

	"pz-web-backend/internal/application/updateapp"
	sysupdate "pz-web-backend/internal/system/update"
)

var (
	defaultUpdateOnce sync.Once
	defaultUpdateSvc  updateapp.Service
)

// Deprecated: CheckUpdateLegacy 仅用于兼容旧调用路径。
// 新代码优先通过 internal/application/updateapp。
func CheckUpdateLegacy(githubRepo string, currentVersion string, devMode bool) (string, string, error) {
	return defaultUpdate(githubRepo, currentVersion, devMode).CheckUpdate()
}

// Deprecated: PerformUpdateLegacy 仅用于兼容旧调用路径。
// 新代码优先通过 internal/application/updateapp。
func PerformUpdateLegacy(githubRepo string, currentVersion string, devMode bool, downloadUrl string) error {
	return defaultUpdate(githubRepo, currentVersion, devMode).PerformUpdate(downloadUrl)
}

func defaultUpdate(githubRepo string, currentVersion string, devMode bool) updateapp.Service {
	defaultUpdateOnce.Do(func() {
		checker := sysupdate.Service{
			HTTPClient:     &http.Client{Timeout: 10 * time.Second},
			GithubRepo:     githubRepo,
			CurrentVersion: currentVersion,
		}
		defaultUpdateSvc = updateapp.NewService(devMode, checker)
	})
	return defaultUpdateSvc
}
