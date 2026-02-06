package mods

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

type WorkshopClient struct {
	httpClient *http.Client
	apiURL     string
	cache      CacheStore

	mu    sync.RWMutex
	mem   map[string]ModInfo
	dirty bool
}

func NewWorkshopClient(httpClient *http.Client, apiURL string, cache CacheStore) (*WorkshopClient, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if apiURL == "" {
		apiURL = "https://api.steampowered.com/ISteamRemoteStorage/GetPublishedFileDetails/v1/"
	}

	c := &WorkshopClient{
		httpClient: httpClient,
		apiURL:     apiURL,
		cache:      cache,
		mem:        make(map[string]ModInfo),
	}

	if cache != nil {
		if loaded, err := cache.Load(); err == nil && loaded != nil {
			c.mem = loaded
		}
	}
	return c, nil
}

type steamResponse struct {
	Response struct {
		ResultCount          int `json:"resultcount"`
		PublishedFileDetails []struct {
			PublishedFileID string `json:"publishedfileid"`
			Title           string `json:"title"`
			Description     string `json:"description"`
		} `json:"publishedfiledetails"`
	} `json:"response"`
}

func (c *WorkshopClient) FetchWorkshopInfo(workshopID string) (ModInfo, error) {
	c.mu.RLock()
	if info, ok := c.mem[workshopID]; ok {
		c.mu.RUnlock()
		return info, nil
	}
	c.mu.RUnlock()

	data := url.Values{}
	data.Set("itemcount", "1")
	data.Set("publishedfileids[0]", workshopID)

	resp, err := c.httpClient.Post(c.apiURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return ModInfo{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ModInfo{}, err
	}

	var steamResp steamResponse
	if err := json.Unmarshal(body, &steamResp); err != nil {
		return ModInfo{}, err
	}

	if steamResp.Response.ResultCount <= 0 || len(steamResp.Response.PublishedFileDetails) == 0 {
		return ModInfo{}, fmt.Errorf("mod not found")
	}

	details := steamResp.Response.PublishedFileDetails[0]
	modID := extractModID(details.Description)
	if modID == "" {
		modID = "?"
	}

	info := ModInfo{
		Name:        details.Title,
		WorkshopID:  workshopID,
		ModID:       modID,
		Description: details.Description,
	}

	c.mu.Lock()
	c.mem[workshopID] = info
	c.dirty = true
	c.mu.Unlock()

	if c.cache != nil {
		go c.flush()
	}

	return info, nil
}

func (c *WorkshopClient) flush() {
	c.mu.RLock()
	if !c.dirty {
		c.mu.RUnlock()
		return
	}
	snapshot := make(map[string]ModInfo, len(c.mem))
	for k, v := range c.mem {
		snapshot[k] = v
	}
	c.mu.RUnlock()

	_ = c.cache.Save(snapshot)

	c.mu.Lock()
	c.dirty = false
	c.mu.Unlock()
}

func extractModID(desc string) string {
	re := regexp.MustCompile(`(?i)Mod\s*ID\s*:\s*([^\r\n<]+)`)
	matches := re.FindAllStringSubmatch(desc, -1)

	var ids []string
	seen := make(map[string]bool)
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		id := strings.TrimSpace(m[1])
		id = strings.Trim(id, "[]")
		if id == "" || seen[id] {
			continue
		}
		ids = append(ids, id)
		seen[id] = true
	}

	if len(ids) == 0 {
		return ""
	}
	return strings.Join(ids, ",")
}
