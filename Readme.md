# Project Zomboid Web Configurator (PZ-Web)

一个为 [Project Zomboid Dedicated Server](https://store.steampowered.com/app/380870/Project_Zomboid_Dedicated_Server/) 设计的轻量级、Web 可视化配置管理工具。

该项目作为 Docker 容器中的 Sidecar 服务运行，提供了 Web 界面来管理 `servertest.ini` 和 `SandboxVars.lua`，集成了模组管理、多语言支持和服务器控制功能。

---

## ✨ 核心功能

*   **可视化配置编辑**：
    *   自动解析 `Server.ini` 和 `SandboxVars.lua`。
    *   **I18n 支持**：直接读取游戏原生翻译文件，自动显示配置项的中文/英文名称和 Tooltip（没翻译的我就没做:p）。
    *   **智能分类**：自动将几百个配置项归类（如“僵尸特性”、“物资稀有度”）。
    *   **表单控件**：自动识别下拉选项（Select）和文本输入（Input）。
    *   **TODO**：将bool变成switch组件，数字类的Option用inputNumber组件替代。

*   **模组管理器**：
    *   **创意工坊集成**：支持直接输入 Workshop ID，自动从 Steam API 获取模组名称。
    *   **智能解析**：自动处理 Workshop ID 与 Mod ID 的对应关系。
    *   **一键应用**：自动生成分号分隔的配置字符串并去重。

*   **服务器监控与控制**：
    *   实时查看 Supervisor 控制台日志。
    *   提供“重启”和“更新并重启”功能（自动触发 SteamCMD 更新）。
    *   提供面板自重启功能，方便build调试

*   **轻量**：
    *   基于 Go (Gin) 编写，编译后仅几 MB。
    *   前端使用 Alpine.js + Tailwind CSS，无 Node.js 依赖，单文件部署。
---

## 🛠️ 开发环境搭建

推荐在 **Windows (WSL2)** 或 **Linux** 环境下进行开发。

### 准备环境
确保已安装 Go 1.23+（`go.mod` 以 Go 1.23 为基线）。

```bash
# 配置 Go 国内代理 (如果你在中国)
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct

# 开启开发模式 (使用 mock 路径验证文件改动)
export DEV_MODE=true 

# 启动服务
go run .
```

---

## 🧱 目录结构（DDD）

后端按模块拆分为 DDD 风格结构（核心逻辑进入 `internal/`，Web 层保持薄）：  

- `internal/application/*`：用例层（Config / I18n / Mods / Update），协调领域逻辑与基础设施
- `internal/config`：`servertest.ini` / `SandboxVars.lua` 的解析与生成（含分组推断、Lua 值格式化）
- `internal/i18n`：读取游戏翻译文件（`lua/shared/Translate`）并提供翻译查询（含资源表）
- `internal/mods`：本地 Workshop 扫描 + Steam Workshop 元信息抓取（含文件缓存）
- `internal/system/update`：GitHub Release 更新检查（checker）
- `internal/infra/*`：副作用与系统依赖（路径推断 / 进程与文件操作等）
- `internal/legacy`：历史兼容入口（Deprecated，仅为重构期间过渡保留）
- `template/`：前端模板与静态资源（见下一节）

---

## 🧩 Template 结构（模块化）

`index.html` 仅保留入口，主体按模块拆分：

- `template/index.html`：入口（引入脚本、挂载 Alpine `x-data="app()"`）
- `template/app_body.html`：页面组装（仅引用各模块模板）
- `template/partials/*`：通用 UI 片段（导航栏 / Tabs / Toast / Mobile Dock）
- `template/modules/*`：业务模块 UI（Server / Sandbox / Monitor / Mod Modal）
- `template/assets/app/main.js`：前端主逻辑（从原内联脚本迁移）

说明：
- 模板加载使用 `template_loader.go` 遍历 `template/` 下所有 `.html`，避免 `ParseFS` 不支持 `**` 递归通配符的问题。

---

## 🧪 单元测试

```bash
go test ./...
```

---

## 🧷 DEV_MODE / Mock 数据

开发模式会优先使用 `testdata/`（Go 约定的测试夹具目录）：

- `testdata/mock_zomboid`：最小化的配置文件夹具
- `testdata/mock_media`：最小化的翻译文件夹具（仅覆盖测试/演示所需的少量 Key）

如果你希望在开发模式下看到完整翻译与资源（体积较大），使用：

- `testdata/mock_zomboid_full`
- `testdata/mock_media_full`

本地 Workshop 扫描（`/api/mods`）在开发模式默认读取：

- `testdata/pzserver/steamapps/...`（结构对齐 `/opt/pzserver/steamapps/...`）
