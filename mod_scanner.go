package main

import (
	"bufio"
	"fmt"
	"io/fs"
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

// 给定安装目录，找到所有 mod.info
func ScanLocalMods(installDir string) ([]ModInfo, error) {
	if installDir == "" {
		if os.Getenv("DEV_MODE") == "true" {
			installDir = "./"
		} else {
			installDir = "/opt/pzserver"
		}
	}
	var mods []ModInfo

	// 基础路径: /opt/pzserver/steamapps/workshop/content/108600
	workshopBase := filepath.Join(installDir, "steamapps", "workshop", "content", "108600")

	// 检查根目录
	if _, err := os.Stat(workshopBase); os.IsNotExist(err) {
		return mods, fmt.Errorf("workshop base not found: %s", workshopBase)
	}

	fmt.Printf("[Scanner] Deep scanning in: %s\n", workshopBase)

	// 使用 WalkDir 进行递归遍历 (比 Walk 更快)
	err := filepath.WalkDir(workshopBase, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // 跳过错误，继续扫描其他文件
		}

		// 我们只关心名叫 mod.info 的文件
		if !d.IsDir() && strings.EqualFold(d.Name(), "mod.info") {

			// 推断 WorkshopID
			// 路径通常是: .../108600/{WorkshopID}/mods/{ModName}/{Version?}/mod.info
			// 往上找，找到数字命名的那个目录，就是 WorkshopID
			wsID := extractWorkshopID(path, workshopBase)

			// 解析文件
			info, parseErr := parseModInfo(path, wsID)
			if parseErr == nil {
				// 去重
				// 如果同一个 WorkshopID 下有多个版本的 mod.info (比如 42 和 42.13)
				// 它们的 ModID 通常是一样的,需要决定保留哪个。
				// 简单策略：如果 ModID 已存在，覆盖它（假设后遍历到的是深层/新版）
				// 直接添加到列表，但在添加前检查一下是否重复
				found := false
				for i, existing := range mods {
					if existing.ModID == info.ModID && existing.WorkshopID == info.WorkshopID {
						// 发现重复。
						// 这通常意味着有多版本文件夹 (mods/X/42/mod.info 和 mods/X/mod.info)
						// 保留路径更长的那个（通常是特定版本文件夹）
						// 简单策略：保留后发现的
						mods[i] = info
						found = true
						break
					}
				}
				if !found {
					mods = append(mods, info)
				}
			}
		}
		return nil
	})

	fmt.Printf("[Scanner] Scan complete. Found %d mods.\n", len(mods))
	return mods, err
}

// 从路径中提取 WorkshopID
// 路径示例: /opt/.../108600/123456/mods/ModA/mod.info -> 返回 123456
func extractWorkshopID(fullPath string, basePath string) string {
	// 去掉基础路径前缀
	rel, err := filepath.Rel(basePath, fullPath)
	if err != nil {
		return "?"
	}

	// 分割路径
	parts := strings.Split(rel, string(os.PathSeparator))

	// 第一个部分通常就是 WorkshopID (数字)
	if len(parts) > 0 {
		return parts[0]
	}
	return "?"
}

// 解析 mod.info
func parseModInfo(path string, wsID string) (ModInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return ModInfo{}, err
	}
	defer file.Close()

	info := ModInfo{WorkshopID: wsID}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		// 处理 UTF-8 BOM 头 (有些文件开头会有乱码)
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
