# Project Zomboid Web Configurator (PZ-Web)

一个为 [Project Zomboid Dedicated Server](https://store.steampowered.com/app/380870/Project_Zomboid_Dedicated_Server/) 设计的轻量级、Web 可视化配置管理工具。

该项目作为 Docker 容器中的 Sidecar 服务运行，提供了一个现代化的 Web 界面来管理复杂的 `servertest.ini` 和 `SandboxVars.lua`，并集成了模组管理、多语言支持和服务器控制功能。

---

## ✨ 核心功能

*   **可视化配置编辑**：
    *   自动解析 `Server.ini` 和 `SandboxVars.lua`。
    *   **I18n 支持**：直接读取游戏原生翻译文件，自动显示配置项的中文/英文名称和 Tooltip。
    *   **智能分类**：自动将几百个配置项归类（如“僵尸特性”、“物资稀有度”）。
    *   **表单控件**：自动识别下拉选项（Select）和文本输入（Input）。

*   **模组管理器**：
    *   **创意工坊集成**：支持直接输入 Workshop ID，自动从 Steam API 获取模组名称。
    *   **智能解析**：自动处理 Workshop ID 与 Mod ID 的对应关系。
    *   **一键应用**：自动生成分号分隔的配置字符串并去重。

*   **服务器监控与控制**：
    *   实时查看 Supervisor 控制台日志。
    *   提供“重启”和“更新并重启”功能（自动触发 SteamCMD 更新）。

*   **轻量**：
    *   基于 Go (Gin) 编写，编译后仅几 MB。
    *   前端使用 Alpine.js + Tailwind CSS，无 Node.js 依赖，单文件部署。
---

## 🛠️ 开发环境搭建

推荐在 **Windows (WSL2)** 或 **Linux** 环境下进行开发。

### 准备环境
确保已安装 Go 1.20+。

```bash
# 配置 Go 国内代理 (如果你在中国)
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct

# 开启开发模式 (使用 mock 路径验证文件改动)
export DEV_MODE=true 

# 启动服务
go run .
```
