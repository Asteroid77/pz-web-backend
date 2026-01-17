package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/gin-gonic/gin"
)

type GithubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadUrl string `json:"browser_download_url"`
	} `json:"assets"`
}

// CheckUpdate 检查 GitHub 最新版
func CheckUpdate() (string, string, error) {
	// 构造 API URL (注意：GithubRepo 必须是 "user/repo" 格式)
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", GithubRepo)

	client := http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "PZ-Web-Configurator")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		// 读取限流重置时间
		resetTimeStr := resp.Header.Get("X-RateLimit-Reset")
		limit := resp.Header.Get("X-RateLimit-Limit")
		remaining := resp.Header.Get("X-RateLimit-Remaining")

		errMsg := "github api rate limit exceeded"
		if resetTimeStr != "" {
			// 转换时间戳
			ts, _ := strconv.ParseInt(resetTimeStr, 10, 64)
			resetTime := time.Unix(ts, 0)
			errMsg = fmt.Sprintf("GitHub API rate limitation (%s/%s). retry it after %s",
				remaining, limit, resetTime.Format("15:04:05"))
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

	// 版本对比 (使用 SemVer)
	// 如果当前版本是 "dev" 或空，我们通常不建议自动更新，或者总是提示更新
	// 这里假设 "dev" 永远不更新，除非强制
	if Version == "dev" {
		return "", "", nil
	}

	// 解析版本号 (自动处理 v 前缀)
	vCurrent, err := semver.NewVersion(Version)
	if err != nil {
		// 如果当前版本号格式不对 (比如 commit hash)，没法比，就跳过
		return "", "", nil
	}

	vLatest, err := semver.NewVersion(release.TagName)
	if err != nil {
		return "", "", fmt.Errorf("invalid release tag: %s", release.TagName)
	}

	// 只有当 Latest > Current 时才提示更新
	if vLatest.GreaterThan(vCurrent) {
		// 寻找对应架构的二进制文件
		expectedPrefix := "pz-web-backend"
		osName := runtime.GOOS     // linux, windows, darwin
		archName := runtime.GOARCH // amd64, arm64
		fmt.Printf("[Update] Looking for asset with: %s, %s, %s\n", expectedPrefix, osName, archName)

		for _, asset := range release.Assets {
			name := strings.ToLower(asset.Name)

			// 使用模糊匹配 (包含 OS 和 Arch 即可)
			if strings.Contains(name, expectedPrefix) &&
				strings.Contains(name, osName) &&
				strings.Contains(name, archName) {
				return release.TagName, asset.BrowserDownloadUrl, nil
			}
		}

		return "", "", fmt.Errorf("no binary found for %s/%s in release %s", osName, archName, release.TagName)
	}

	return "", "", nil // 已是最新
}

// PerformUpdate 执行下载和替换
func PerformUpdate(downloadUrl string) error {
	// 下载新文件
	resp, err := http.Get(downloadUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	tmpPath := "/tmp/pz-config-app.new"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	io.Copy(out, resp.Body)
	out.Close()

	// 授予执行权限
	os.Chmod(tmpPath, 0755)

	// 替换旧文件 (原子操作)
	// 在 Linux 中，即使程序正在运行，也可以重命名它的二进制文件
	binPath, err := os.Executable()

	if err != nil {
		//怎么会获取不到呢，真获取不到就硬编码
		binPath = "/opt/pz-web-backend/pz-web-backend"
	}

	fmt.Printf("[Update] Target binary path: %s\n", binPath)

	// 如果是开发模式，不能把自己覆盖
	if os.Getenv("DEV_MODE") == "true" {
		fmt.Println("[Mock] Update file downloaded to", tmpPath)
		return nil
	}

	// 备份旧文件
	os.Rename(binPath, binPath+".bak")

	// 移动新文件到位
	err = os.Rename(tmpPath, binPath)
	if err != nil {
		// 失败回滚
		os.Rename(binPath+".bak", binPath)
		return err
	}

	// 触发重启
	// 使用 nohup 异步重启，防止当前 HTTP 请求被中断导致客户端收不到响应
	// 或者直接让前端由 timeout 处理
	go func() {
		cmd := exec.Command("supervisorctl", "restart", "webconfig")
		cmd.Run()
	}()

	return nil
}
func handleRestartPanel(c *gin.Context) {
	// 异步执行，防止 HTTP 请求被中断导致前端报错
	go func() {
		// 稍微延迟一下，给 HTTP 响应一点时间返回
		time.Sleep(1 * time.Second)

		// 这里的 webconfig 必须和你 supervisor conf 里的 program 名称一致
		cmd := exec.Command("supervisorctl", "restart", "webconfig")
		if err := cmd.Run(); err != nil {
			fmt.Println("Restart failed:", err)
		}
	}()

	c.JSON(200, gin.H{"status": "ok", "message": "Restaring..."})
}
