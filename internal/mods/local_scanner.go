package mods

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ScanLocalMods 扫描 Steam Workshop 本地已安装模组（通过 mod.info）。
//
// installDir:
// - 为空时由调用方决定默认值（避免在此层依赖环境变量与运行目录）。
func ScanLocalMods(installDir string) ([]ModInfo, error) {
	var mods []ModInfo

	workshopBase := filepath.Join(installDir, "steamapps", "workshop", "content", "108600")
	if _, err := os.Stat(workshopBase); os.IsNotExist(err) {
		return mods, fmt.Errorf("workshop base not found: %s", workshopBase)
	}

	err := filepath.WalkDir(workshopBase, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || !strings.EqualFold(d.Name(), "mod.info") {
			return nil
		}

		wsID := extractWorkshopID(path, workshopBase)
		info, parseErr := parseModInfo(path, wsID)
		if parseErr != nil {
			return nil
		}

		found := false
		for i, existing := range mods {
			if existing.ModID == info.ModID && existing.WorkshopID == info.WorkshopID {
				mods[i] = info
				found = true
				break
			}
		}
		if !found {
			mods = append(mods, info)
		}

		return nil
	})

	return mods, err
}

func extractWorkshopID(fullPath string, basePath string) string {
	rel, err := filepath.Rel(basePath, fullPath)
	if err != nil {
		return "?"
	}
	parts := strings.Split(rel, string(os.PathSeparator))
	if len(parts) > 0 {
		return parts[0]
	}
	return "?"
}

func parseModInfo(path string, wsID string) (ModInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return ModInfo{}, err
	}
	defer file.Close()

	info := ModInfo{WorkshopID: wsID}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		text = strings.TrimPrefix(text, "\uFEFF")
		if text == "" || strings.HasPrefix(text, "//") {
			continue
		}

		parts := strings.SplitN(text, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			info.Name = val
		case "id":
			info.ModID = val
		case "description":
			info.Description = val
		}
	}

	if info.ModID == "" {
		return info, fmt.Errorf("no id")
	}
	if info.Name == "" {
		info.Name = info.ModID
	}

	return info, nil
}
