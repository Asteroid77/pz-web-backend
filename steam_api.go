package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
)

type SteamResponse struct {
	Response struct {
		ResultCount          int `json:"resultcount"`
		PublishedFileDetails []struct {
			PublishedFileID string `json:"publishedfileid"`
			Title           string `json:"title"`
			Description     string `json:"description"`
		} `json:"publishedfiledetails"`
	} `json:"response"`
}

// 缓存文件路径 (在 WSL/Docker 里存在 /opt/pz-web-backend/ 目录下)
const CacheFilePath = "/opt/pz-web-backend/workshop_cache.json"

// 简单的内存缓存，防止重复请求 Steam
var steamCache = make(map[string]ModInfo)
var steamMutex sync.RWMutex

// 初始化时加载缓存文件
func init() {
	loadCacheFromFile()
}

// 从文件加载缓存
func loadCacheFromFile() {
	path := CacheFilePath
	// 如果是本地开发模式，改路径为当前目录下
	if os.Getenv("DEV_MODE") == "true" {
		path = "workshop_cache.json"
	}

	file, err := os.Open(path)
	if err != nil {
		return // 文件不存在则忽略
	}
	defer file.Close()

	bytes, _ := io.ReadAll(file)

	steamMutex.Lock()
	defer steamMutex.Unlock()
	json.Unmarshal(bytes, &steamCache)
	fmt.Printf("Loaded %d mod info from cache\n", len(steamCache))
}

// 保存缓存到文件
func saveCacheToFile() {
	path := CacheFilePath
	if os.Getenv("DEV_MODE") == "true" {
		path = "workshop_cache.json"
	}

	steamMutex.RLock()
	data, err := json.MarshalIndent(steamCache, "", "  ")
	steamMutex.RUnlock()

	if err == nil {
		os.WriteFile(path, data, 0644)
	}
}

func extractModID(desc string) string {
	// 匹配模式: Mod ID: xxx (忽略大小写)
	// 有些描述里有多个 Mod ID，用逗号分隔
	re := regexp.MustCompile(`(?i)Mod\s*ID\s*:\s*([^\r\n<]+)`)
	matches := re.FindAllStringSubmatch(desc, -1)

	var ids []string
	seen := make(map[string]bool)

	for _, m := range matches {
		if len(m) >= 2 {
			id := strings.TrimSpace(m[1])
			// 去除可能残留的 HTML 标签或多余字符
			id = strings.Trim(id, "[]")
			if id != "" && !seen[id] {
				ids = append(ids, id)
				seen[id] = true
			}
		}
	}

	if len(ids) > 0 {
		return strings.Join(ids, ",")
	}
	return ""
}

func FetchWorkshopInfo(workshopID string) (ModInfo, error) {
	// 查缓存
	steamMutex.RLock()
	if info, ok := steamCache[workshopID]; ok {
		steamMutex.RUnlock()
		return info, nil
	}
	steamMutex.RUnlock()

	apiURL := "https://api.steampowered.com/ISteamRemoteStorage/GetPublishedFileDetails/v1/"

	data := url.Values{}
	data.Set("itemcount", "1")
	data.Set("publishedfileids[0]", workshopID)

	resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return ModInfo{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ModInfo{}, err
	}

	var steamResp SteamResponse
	err = json.Unmarshal(body, &steamResp)
	if err != nil {
		return ModInfo{}, err
	}

	if steamResp.Response.ResultCount > 0 && len(steamResp.Response.PublishedFileDetails) > 0 {
		details := steamResp.Response.PublishedFileDetails[0]
		modID := extractModID(details.Description)
		if modID == "" {
			modID = "?" // 没提取到，让用户手动填
		}

		info := ModInfo{
			Name:        details.Title,
			WorkshopID:  workshopID,
			ModID:       modID,
			Description: details.Description,
		}
		// 写入缓存
		steamMutex.Lock()
		steamCache[workshopID] = info
		steamMutex.Unlock()
		// 异步持久化，防止阻塞请求
		go saveCacheToFile()
		// 返回结果
		return info, nil
	}

	return ModInfo{}, fmt.Errorf("mod not found")
}
