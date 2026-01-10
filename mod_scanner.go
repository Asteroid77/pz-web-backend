package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ModInfo struct {
	Name        string `json:"name"`
	ModID       string `json:"mod_id"`
	WorkshopID  string `json:"workshop_id"`
	Description string `json:"description"`
}

// 假设 Steam 的创意工坊目录
// 在 Dockerfile 里我们是 steam 用户，路径通常是 /home/steam/steamapps/workshop/content/108600
var WorkshopContentDir = "/home/steam/steamapps/workshop/content/108600"

func init() {
	if os.Getenv("DEV_MODE") == "true" {
		cwd, _ := os.Getwd()
		WorkshopContentDir = filepath.Join(cwd, "mock_workshop")
	}
}

// ScanInstalledMods 扫描已下载的所有 Mod
func ScanInstalledMods() ([]ModInfo, error) {
	var mods []ModInfo

	// 遍历 WorkshopID 目录 (如 108600/2857548524)
	workshopDirs, err := os.ReadDir(WorkshopContentDir)
	if err != nil {
		return nil, err
	}

	for _, wsDir := range workshopDirs {
		if !wsDir.IsDir() {
			continue
		}
		wsID := wsDir.Name()

		// 进入 mods 目录 (如 108600/2857548524/mods)
		modsPath := filepath.Join(WorkshopContentDir, wsID, "mods")
		modDirs, err := os.ReadDir(modsPath)
		if err != nil {
			// 有些 workshop item 可能不是 mod 或者是旧结构，尝试直接找 mod.info
			// 简单起见，这里假设标准结构
			continue
		}

		for _, mDir := range modDirs {
			if !mDir.IsDir() {
				continue
			}
			// 读取 mod.info
			infoPath := filepath.Join(modsPath, mDir.Name(), "mod.info")
			info, err := parseModInfo(infoPath)
			if err == nil {
				info.WorkshopID = wsID
				steamMutex.RLock()
				if cached, ok := steamCache[wsID]; ok {
					// 拼接 (例如: "Common Sense ([B41] Common Sense)")
					info.Name = fmt.Sprintf("%s (%s)", cached.Name, info.Name)
				}
				steamMutex.RUnlock()
				mods = append(mods, info)
			}
		}
	}
	return mods, nil
}

// 解析 mod.info 文件
func parseModInfo(path string) (ModInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return ModInfo{}, err
	}
	defer f.Close()

	info := ModInfo{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.ToLower(parts[0])
			val := parts[1]
			switch key {
			case "name":
				info.Name = val
			case "id":
				info.ModID = val
			case "description":
				info.Description = val
			}
		}
	}
	return info, nil
}
