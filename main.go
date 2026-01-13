package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	BaseDataDir string                        // 存档目录 (读取用户配置)
	BaseGameDir string                        // 游戏目录 (读取官方翻译)
	Version     = "dev"                       // 默认值 (开发模式显示 dev)
	CommitSHA   = "unknown"                   // Git Commit Hash
	BuildTime   = "unknown"                   // 构建时间
	GithubRepo  = "Asteroid77/pz-web-backend" // GitHub 仓库 (user/repo)
)

//go:embed template
var contentFS embed.FS

func init() {
	if os.Getenv("DEV_MODE") == "true" {
		cwd, _ := os.Getwd()
		BaseDataDir = filepath.Join(cwd, "mock_zomboid")
		BaseGameDir = filepath.Join(cwd, "mock_media")
		fmt.Println("Running in DEV MODE")
	} else {
		BaseDataDir = os.Getenv("PZ_DATA_DIR")
		if BaseDataDir == "" {
			BaseDataDir = "/home/steam/Zomboid" // 默认值
		}

		// 读取 PZ_INSTALL_DIR (对应 /opt/pzserver)
		installDir := os.Getenv("PZ_INSTALL_DIR")
		if installDir == "" {
			installDir = "/opt/pzserver" // 默认值
		}
		// 拼装 media 路径
		BaseGameDir = filepath.Join(installDir, "media")

		fmt.Printf("Config Loaded: DataDir=%s, GameDir=%s\n", BaseDataDir, BaseGameDir)
	}
}

func main() {
	fs.WalkDir(contentFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		fmt.Println("Found:", path)
		return nil
	})
	r := gin.Default()
	assetsFS, _ := fs.Sub(contentFS, "template/assets")
	r.StaticFS("/assets", http.FS(assetsFS))
	tmpl := template.Must(template.New("").ParseFS(contentFS, "template/*.html"))
	r.SetHTMLTemplate(tmpl)

	// --- API 路由 ---
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "PZ Server Manager",
		})
	})
	// 获取 servertest.ini
	r.GET("/api/config/server", func(c *gin.Context) {
		// 获取 URL 参数 ?lang=CN，默认值为 CN
		lang := c.DefaultQuery("lang", "CN")
		// 将 lang 转为大写，以防用户传 cn
		lang = strings.ToUpper(lang)
		path := filepath.Join(BaseDataDir, "Server/servertest.ini")
		items, err := ParseServerTestIni(path, lang)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"filename": "servertest.ini", "items": items})
	})

	// 获取 SandboxVars.lua
	r.GET("/api/config/sandbox", func(c *gin.Context) {
		// 获取 URL 参数 ?lang=CN，默认值为 CN
		lang := c.DefaultQuery("lang", "CN")
		// 将 lang 转为大写，以防用户传 cn
		lang = strings.ToUpper(lang)

		path := filepath.Join(BaseDataDir, "Server/servertest_SandboxVars.lua")

		// 传入 lang
		items, err := ParseSandboxLua(path, lang)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"filename": "SandboxVars.lua", "lang": lang, "items": items})
	})
	r.GET("/api/logs", func(c *gin.Context) {
		// 为了安全和简单，这里演示读取 /var/log/supervisord.log
		// 实际可以用 cmd 输出
		path := "/home/steam/pz-stdout.log"
		f, err := os.Open(path)
		if err != nil {
			// 如果文件不存在（比如刚启动还没产生日志），返回提示
			c.JSON(200, gin.H{"logs": "等待服务启动，日志暂为空..."})
			return
		}
		defer f.Close()
		// 获取文件大小
		stat, _ := f.Stat()
		fileSize := stat.Size()

		// 只读取最后 10KB (约5000-10000字)，避免把浏览器卡死
		// 如果你想看更多，可以调大这个值
		const readSize = 10240

		var content []byte
		if fileSize > readSize {
			// 如果文件很大，从末尾开始读
			content = make([]byte, readSize)
			// 移动光标到 倒数 readSize 的位置
			f.Seek(-readSize, 2)
			f.Read(content)
		} else {
			// 如果文件小，全读
			content, _ = os.ReadFile(path)
		}

		c.JSON(200, gin.H{"logs": string(content)})
	})
	r.POST("/api/action/update_restart", func(c *gin.Context) {
		// 重启 pzserver。
		c.JSON(200, gin.H{"status": "ok", "message": "正在后台执行更新与重启..."})

		// 2. 异步执行
		go func() {
			fmt.Println(">>> 准备执行重启命令...")

			// 使用绝对路径 /usr/bin/supervisorctl
			// -c 指定配置文件，确保它能找到 socket 路径
			cmd := exec.Command("/usr/bin/supervisorctl", "-c", "/etc/supervisor/conf.d/supervisord.conf", "restart", "pzserver")

			// 捕获所有输出（包括错误信息）
			output, err := cmd.CombinedOutput()

			if err != nil {
				fmt.Printf("!!! 重启执行失败 !!!\nError: %v\nOutput: %s\n", err, string(output))
			} else {
				fmt.Printf(">>> 重启指令执行成功:\n%s\n", string(output))
			}
		}()
	})
	r.GET("/api/i18n", func(c *gin.Context) {
		currentLang := strings.ToUpper(c.DefaultQuery("lang", "CN"))

		// 1. 扫描 Translate 目录
		translateDir := filepath.Join(BaseGameDir, "lua/shared/Translate")
		files, _ := os.ReadDir(translateDir)

		var languages []LanguageOption

		for _, f := range files {
			if f.IsDir() {
				code := strings.ToUpper(f.Name())
				// 获取显示名称，如果没有则用 Code
				name, ok := LangNameMap[code]
				if !ok {
					name = code
				}
				languages = append(languages, LanguageOption{Code: code, Name: name})
			}
		}

		// 排序：把 CN 和 EN 放在最前面
		sort.Slice(languages, func(i, j int) bool {
			if languages[i].Code == "CN" {
				return true
			}
			if languages[j].Code == "CN" {
				return false
			}
			if languages[i].Code == "EN" {
				return true
			}
			if languages[j].Code == "EN" {
				return false
			}
			return languages[i].Code < languages[j].Code
		})

		// 2. 获取当前语言的 UI 文本
		// 如果请求的语言没有 UI 翻译（比如 AR），回退到 EN
		uiResources, ok := WebUIResources[currentLang]
		if !ok {
			uiResources = WebUIResources["EN"]
		}

		c.JSON(200, gin.H{
			"languages": languages,
			"ui":        uiResources,
		})
	})
	// 根据模组id查找模组信息
	r.GET("/api/mods/lookup", func(c *gin.Context) {
		idsStr := c.Query("ids")
		if idsStr == "" {
			c.JSON(400, gin.H{"error": "ids required"})
			return
		}

		targetIds := strings.Split(idsStr, ",")
		var results []ModInfo

		// 1. 先获取所有本地已安装的
		localMods, _ := ScanLocalMods("")

		for _, wid := range targetIds {
			found := false
			// 尝试在本地找
			for _, lm := range localMods {
				if lm.WorkshopID == wid {
					results = append(results, lm)
					found = true
					// 注意：一个 WorkshopID 可能对应多个 ModID，这里简单起见全加进去
				}
			}

			// 本地没找到，去爬 Steam
			if !found {
				info, err := FetchWorkshopInfo(wid)
				if err == nil {
					results = append(results, info)
				} else {
					// 失败了也返回一个占位符
					results = append(results, ModInfo{
						Name:       "Network Error / Invalid ID",
						WorkshopID: wid,
						ModID:      "?",
					})
				}
			}
		}

		c.JSON(200, results)
	})
	// 返回当前所有本地模组
	r.GET("/api/mods", func(c *gin.Context) {
		localMods, _ := ScanLocalMods("")
		if localMods == nil {
			c.JSON(200, []ModInfo{})
		}
		c.JSON(200, localMods)
	})
	// 保存配置文件
	r.POST("/api/config/:name", func(c *gin.Context) {
		name := c.Param("name")
		// 确定文件路径
		var path string
		switch name {
		case "server":
			path = filepath.Join(BaseDataDir, "Server/servertest.ini")
		case "sandbox":
			path = filepath.Join(BaseDataDir, "Server/servertest_SandboxVars.lua")
		default:
			c.JSON(400, gin.H{"error": "Invalid config type"})
			return
		}

		var req SaveRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// 转换内容
		var content string
		if name == "server" {
			content = GenerateServerTestIni(req.Items)
		} else {
			content = GenerateSandboxLua(req.Items)
		}

		// 写入文件
		// 确保目录存在
		os.MkdirAll(filepath.Dir(path), 0755)

		// 写入
		err := os.WriteFile(path, []byte(content), 0644)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to write file: " + err.Error()})
			return
		}

		// 给予权限 (防止 root 运行导致的权限问题)
		// 生产环境里 exec.Command("chown", "steam:steam", path).Run()
		if os.Getenv("DEV_MODE") != "true" {
			exec.Command("chown", "steam:steam", path).Run()
		}

		if req.Restart {
			// 异步执行重启，避免阻塞 HTTP 响应
			go func() {
				fmt.Println(">>> [Auto-Restart] 配置保存触发重启...")
				// 使用绝对路径，确保能找到命令
				cmd := exec.Command("/usr/bin/supervisorctl", "-c", "/etc/supervisor/conf.d/supervisord.conf", "restart", "pzserver")

				output, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Printf("!!! [Auto-Restart] 重启失败 !!!\nError: %v\nOutput: %s\n", err, string(output))
				} else {
					fmt.Printf(">>> [Auto-Restart] 重启指令执行成功:\n%s\n", string(output))
				}
			}()

			c.JSON(200, gin.H{"status": "saved_and_restarting", "message": "配置已保存，服务器正在重启..."})
		} else {
			// 不需要重启，仅保存
			c.JSON(200, gin.H{"status": "saved", "message": "配置已保存"})
		}
	})

	// 检查更新
	r.GET("/api/system/check_update", func(c *gin.Context) {
		newVer, url, err := CheckUpdate()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{
			"current":      Version,
			"new_version":  newVer, // 如果为空则无更新
			"download_url": url,
		})
	})

	// 执行更新
	r.POST("/api/system/perform_update", func(c *gin.Context) {
		var req struct {
			Url string `json:"url"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if err := PerformUpdate(req.Url); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "updating"})
	})
	// 启动服务
	r.Run(":10888")
}
