package legacy

import (
	"fmt"
	"path/filepath"

	"pz-web-backend/internal/infra/pzpaths"
	"pz-web-backend/internal/mods"
)

// Deprecated: ScanLocalModsLegacy 仅用于兼容旧调用路径。
// 新代码优先通过 internal/application/modsapp。
func ScanLocalModsLegacy(installDir string, devMode bool) ([]mods.ModInfo, error) {
	if installDir == "" {
		installDir = pzpaths.DefaultInstallDir(devMode)
	}
	fmt.Printf("[Scanner] Deep scanning in: %s\n", filepath.Join(installDir, "steamapps", "workshop", "content", "108600"))
	return mods.ScanLocalMods(installDir)
}
