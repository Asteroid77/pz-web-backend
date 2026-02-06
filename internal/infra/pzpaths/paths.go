package pzpaths

import (
	"os"
	"path/filepath"

	"pz-web-backend/internal/mods"
)

func DefaultInstallDir(devMode bool) string {
	if devMode {
		dir := filepath.Join(".", "testdata", "pzserver")
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
		return "."
	}
	return "/opt/pzserver"
}

func WorkshopCachePath(devMode bool) string {
	if devMode {
		return "workshop_cache.json"
	}
	return mods.DefaultCacheFilePath
}
