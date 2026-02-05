package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

type GithubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadUrl string `json:"browser_download_url"`
	} `json:"assets"`
}

type Service struct {
	HTTPClient     *http.Client
	GithubRepo     string
	CurrentVersion string

	OS   string
	Arch string

	// APIBase 允许在测试中替换为 httptest server。
	// 默认: https://api.github.com
	APIBase string
}

func (s Service) CheckUpdate() (string, string, error) {
	client := s.httpClient()

	apiBase := s.APIBase
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}

	repo := strings.TrimSpace(s.GithubRepo)
	if repo == "" {
		return "", "", fmt.Errorf("github repo is empty")
	}

	url := fmt.Sprintf("%s/repos/%s/releases/latest", strings.TrimRight(apiBase, "/"), repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "PZ-Web-Configurator")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		resetTimeStr := resp.Header.Get("X-RateLimit-Reset")
		limit := resp.Header.Get("X-RateLimit-Limit")
		remaining := resp.Header.Get("X-RateLimit-Remaining")

		errMsg := "github api rate limit exceeded"
		if resetTimeStr != "" {
			if ts, parseErr := parseUnix(resetTimeStr); parseErr == nil {
				resetTime := time.Unix(ts, 0)
				errMsg = fmt.Sprintf("GitHub API rate limitation (%s/%s). retry it after %s",
					remaining, limit, resetTime.Format("15:04:05"))
			}
		}
		return "", "", fmt.Errorf(errMsg)
	}

	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("github api returned %d", resp.StatusCode)
	}

	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}

	if s.CurrentVersion == "" || s.CurrentVersion == "dev" {
		return "", "", nil
	}

	vCurrent, err := semver.NewVersion(s.CurrentVersion)
	if err != nil {
		return "", "", nil
	}

	vLatest, err := semver.NewVersion(release.TagName)
	if err != nil {
		return "", "", fmt.Errorf("invalid release tag: %s", release.TagName)
	}

	if !vLatest.GreaterThan(vCurrent) {
		return "", "", nil
	}

	expectedPrefix := "pz-web-backend"
	osName := s.OS
	if osName == "" {
		osName = runtime.GOOS
	}
	archName := s.Arch
	if archName == "" {
		archName = runtime.GOARCH
	}

	for _, asset := range release.Assets {
		name := strings.ToLower(asset.Name)
		if strings.Contains(name, expectedPrefix) &&
			strings.Contains(name, osName) &&
			strings.Contains(name, archName) {
			return release.TagName, asset.BrowserDownloadUrl, nil
		}
	}

	return "", "", fmt.Errorf("no binary found for %s/%s in release %s", osName, archName, release.TagName)
}

func (s Service) httpClient() *http.Client {
	if s.HTTPClient != nil {
		return s.HTTPClient
	}
	return &http.Client{Timeout: 10 * time.Second}
}

func parseUnix(s string) (int64, error) {
	var n int64
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, fmt.Errorf("not a unix timestamp")
		}
		n = n*10 + int64(s[i]-'0')
	}
	return n, nil
}
