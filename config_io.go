package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// 定义 UI 资源里的分组 Key 常量
const (
	SecGeneral     = "general_settings"
	SecAntiCheat   = "anticheat"
	SecMap         = "map_settings"
	SecMods        = "mods_workshop"
	SecDiscord     = "discord_integration"
	SecPlayers     = "players_pvp"
	SecClient      = "client_limits"
	SecLoot        = "loot_rarity"
	SecTime        = "time_settings"
	SecZombieSpawn = "zombie_population"
	SecWorld       = "world_environment"
	SecVehicles    = "vehicle_settings"
	SecCharXP      = "character_exp"
	SecNature      = "nature_agriculture"
	SecZombieLore  = "zombie_lore"
)

// 安全获取分组名称
// 如果当前语言没有对应的翻译，尝试回退到 EN，如果还没有，直接返回 Key
func getSafeSectionName(lang string, sectionKey string) string {
	// 尝试从当前语言获取
	if langMap, ok := WebUIResources[lang]; ok {
		if val, ok := langMap[sectionKey]; ok && val != "" {
			return val
		}
	}

	// 尝试回退到 EN
	fallbackLang := "EN"

	if langMap, ok := WebUIResources[fallbackLang]; ok {
		if val, ok := langMap[sectionKey]; ok && val != "" {
			return val
		}
	}

	// 实在没有，返回 Key 本身，防止分组为空
	return sectionKey
}

// 对于 Server.ini 根据 Key 猜测分组
func getServerOptionSectionName(key string, language string) string {
	key = strings.ToLower(key)

	// 反作弊相关
	if strings.HasPrefix(key, "anticheat") || strings.Contains(key, "hack") || strings.Contains(key, "checksum") {
		return getSafeSectionName(language, SecAntiCheat)
	}

	// 地图与出生点
	if strings.HasPrefix(key, "map") || strings.HasPrefix(key, "spawn") {
		return getSafeSectionName(language, SecMap)
	}

	// Steam、工坊与模组
	if strings.HasPrefix(key, "steam") || strings.HasPrefix(key, "workshop") || strings.HasPrefix(key, "mod") {
		// 特例：SteamVAC 属于反作弊
		if key == "steamvac" {
			return getSafeSectionName(language, SecAntiCheat)
		}
		// 特例：SteamScoreboard 属于常规或玩家设置，这里暂归常规
		if strings.Contains(key, "scoreboard") {
			return getSafeSectionName(language, SecGeneral)
		}
		return getSafeSectionName(language, SecMods)
	}

	// Discord
	if strings.HasPrefix(key, "discord") {
		return getSafeSectionName(language, SecDiscord)
	}

	// 玩家与 PVP
	if strings.Contains(key, "player") || strings.Contains(key, "pvp") || strings.Contains(key, "safehouse") {
		return getSafeSectionName(language, SecPlayers)
	}

	// 客户端限制
	if strings.Contains(key, "client") || strings.Contains(key, "ping") || strings.Contains(key, "speed") {
		return getSafeSectionName(language, SecClient)
	}

	// 其他归为常规
	return getSafeSectionName(language, SecGeneral)
}

// 对于 Sandbox 根据 Key 猜测分组
func getSandboxSection(fullKey string, language string) string {
	// 拆分 Parent.Key
	parts := strings.Split(fullKey, ".")
	parent := ""
	key := fullKey

	if len(parts) > 1 {
		parent = parts[0]
		key = parts[1]
	}

	// 优先根据父级表名归类
	switch parent {
	case "ZombieLore":
		return getSafeSectionName(language, SecZombieLore)
	case "Map":
		return getSafeSectionName(language, SecMap)
	case "Basement":
		return getSafeSectionName(language, SecMap) // Basement 也可以归类到地图
	case "MultiplierConfig":
		return getSafeSectionName(language, SecCharXP)
	case "ZombieConfig":
		return getSafeSectionName(language, SecZombieSpawn)
	}

	k := strings.ToLower(key)

	// 物资 (Loot) - 通常以 Loot 结尾
	if strings.HasSuffix(k, "loot") || strings.Contains(k, "lootnew") {
		return getSafeSectionName(language, SecLoot)
	}
	// 时间 (Time)
	if strings.Contains(k, "start") || strings.Contains(k, "time") || key == "DayLength" {
		return getSafeSectionName(language, SecTime)
	}
	// 僵尸生成 (Population / Advanced)
	if strings.Contains(k, "zombie") || strings.Contains(k, "respawn") ||
		strings.Contains(k, "pop") || strings.Contains(k, "group") ||
		strings.Contains(k, "redistribute") || key == "Distribution" {
		return getSafeSectionName(language, SecZombieSpawn)
	}
	// 世界 (World)
	if strings.Contains(k, "water") || strings.Contains(k, "elec") ||
		strings.Contains(k, "fire") || strings.Contains(k, "rot") ||
		strings.Contains(k, "alarm") || strings.Contains(k, "house") ||
		strings.Contains(k, "generator") || strings.Contains(k, "fuel") {
		return getSafeSectionName(language, SecWorld)
	}
	// 车辆 (Vehicles)
	if strings.Contains(k, "car") || strings.Contains(k, "traffic") ||
		strings.Contains(k, "gas") || strings.Contains(k, "vehicle") {
		return getSafeSectionName(language, SecVehicles)
	}
	// 角色与经验 (Character / XP)
	if strings.Contains(k, "xp") || strings.Contains(k, "stat") ||
		strings.Contains(k, "regen") || strings.Contains(k, "clothing") ||
		strings.Contains(k, "injury") || strings.Contains(k, "move") ||
		strings.Contains(k, "stamina") || strings.Contains(k, "nutrition") {
		return getSafeSectionName(language, SecCharXP)
	}
	// 自然与农业 (Nature / Farming)
	if strings.Contains(k, "erosion") || strings.Contains(k, "farming") ||
		strings.Contains(k, "nature") || strings.Contains(k, "rain") ||
		strings.Contains(k, "temperature") || strings.Contains(k, "plant") ||
		strings.Contains(k, "compost") {
		return getSafeSectionName(language, SecNature)
	}
	// 地图 (Map)
	if strings.Contains(k, "map") || strings.Contains(k, "chunk") {
		return getSafeSectionName(language, SecMap)
	}

	return getSafeSectionName(language, SecGeneral)
}

// 解析server.ini
func ParseServerTestIni(path string, lang string) ([]ConfigItem, error) {
	dict := GetTranslationMap(lang)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var items []ConfigItem
	scanner := bufio.NewScanner(f)

	// 增加 Buffer 大小，ServerWelcomeMessage 可能会非常长
	// 默认 64k，这里设为 1MB 增加冗余
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行、井号注释、分号注释
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// 限制分割次数为 2，防止 Map=Muldraugh, KY 被错误截断
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			// ServerOption 翻译
			_, tooltip := TranslateKey(dict, key, "UI_ServerOption_")

			// 只有当 Tooltip 为空时，才回退显示 Key
			if tooltip == "" {
				tooltip = key
			}

			// 获取分类
			section := getServerOptionSectionName(key, lang)

			items = append(items, ConfigItem{
				Key:     key,
				Value:   val,
				Label:   tooltip, // Server.ini 的配置项通常用 Tooltip 字段存储了它的中文名
				Section: section,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

// ParseSandboxLua 解析 Lua 文件，支持嵌套表结构
func ParseSandboxLua(path string, lang string) ([]ConfigItem, error) {
	dict := GetTranslationMap(lang)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var items []ConfigItem

	// 正则定义
	// 匹配: Key = {
	reTableStart := regexp.MustCompile(`^\s*([A-Za-z0-9_]+)\s*=\s*\{\s*$`)
	// 匹配: }, 或 }
	reTableEnd := regexp.MustCompile(`^\s*\},?\s*$`)
	// 匹配: Key = Value,
	reValue := regexp.MustCompile(`^\s*([A-Za-z0-9_]+)\s*=\s*(.*),`)

	scanner := bufio.NewScanner(f)
	currentContext := "" // 用于记录当前嵌套的表名 (如 "ZombieLore")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		// 检查是否是子表开始 (例如 "ZombieLore = {")
		// 要排除 "SandboxVars = {" 这一行，因为它不仅是开始也是根
		if matches := reTableStart.FindStringSubmatch(line); len(matches) == 2 {
			tableName := matches[1]
			if tableName != "SandboxVars" {
				currentContext = tableName
				continue
			}
		}

		// 检查是否是子表结束 (例如 "},")
		if reTableEnd.MatchString(line) {
			currentContext = ""
			continue
		}

		// 检查是否是普通赋值
		if matches := reValue.FindStringSubmatch(line); len(matches) == 3 {
			rawKey := matches[1]
			val := matches[2]

			// 清理值：去掉末尾逗号（正则已处理，但为了保险），去掉双引号
			val = strings.TrimSuffix(val, ",")
			val = strings.Trim(val, "\"")

			// 构建唯一 Key
			// 如果在子表中，Key 变成 "ZombieLore.Speed"
			// 如果在根目录，Key 保持 "Zombies"
			fullKey := rawKey
			if currentContext != "" {
				fullKey = currentContext + "." + rawKey
			}

			// 翻译查找：使用 rawKey (例如 Speed) 去查找翻译，而不是 fullKey
			// Sandbox_Speed 是存在的，但 Sandbox_ZombieLore.Speed 可能不存在
			label, tooltip := TranslateKey(dict, rawKey, "Sandbox_")

			// 获取选项 (同样用 rawKey)
			options := GetSandboxOptions(dict, rawKey)

			// 智能判断 Section
			section := getSandboxSection(fullKey, lang)

			items = append(items, ConfigItem{
				Key:     fullKey, // 存储带点的 Key，用于生成时还原结构
				Value:   val,
				Label:   label,
				Tooltip: tooltip,
				Options: options,
				Section: section,
			})
		}
	}
	return items, nil
}

func GenerateServerTestIni(items []ConfigItem) string {
	var sb strings.Builder

	// 为了美观，可以稍微分个组，或者直接按顺序写入
	for _, item := range items {
		// INI 格式: Key=Value
		sb.WriteString(fmt.Sprintf("%s=%s\n", item.Key, item.Value))
	}

	return sb.String()
}

// GenerateSandboxLua 生成 Lua 文件，还原嵌套结构
func GenerateSandboxLua(items []ConfigItem) string {
	var sb strings.Builder

	// 分组存储
	// globals 存储根级配置 (Zombies, DayLength)
	// nestedMap 存储嵌套配置 (ZombieLore -> [Speed, Strength...])
	globals := []ConfigItem{}
	nestedMap := make(map[string][]ConfigItem)

	for _, item := range items {
		if strings.Contains(item.Key, ".") {
			parts := strings.SplitN(item.Key, ".", 2)
			parent := parts[0]
			childKey := parts[1]

			// 创建一个新的 item，把 Key 改回短名 (Speed)，方便生成
			subItem := item
			subItem.Key = childKey
			nestedMap[parent] = append(nestedMap[parent], subItem)
		} else {
			globals = append(globals, item)
		}
	}

	sb.WriteString("SandboxVars = {\n")
	sb.WriteString("    VERSION = 6,\n") // 强制写入版本号，或者从 items 里找

	// 写入根级配置
	for _, item := range globals {
		if item.Key == "VERSION" {
			continue
		} // 防止重复
		val := formatLuaValue(item.Value)
		sb.WriteString(fmt.Sprintf("    %s = %s,\n", item.Key, val))
	}

	// 写入嵌套配置
	// 为了保证生成顺序一致性（Git友好），对 map 的 key 进行排序
	var sortedTables []string
	for k := range nestedMap {
		sortedTables = append(sortedTables, k)
	}
	sort.Strings(sortedTables)

	for _, tableName := range sortedTables {
		subItems := nestedMap[tableName]
		sb.WriteString(fmt.Sprintf("    %s = {\n", tableName))

		for _, item := range subItems {
			val := formatLuaValue(item.Value)
			sb.WriteString(fmt.Sprintf("        %s = %s,\n", item.Key, val))
		}

		sb.WriteString("    },\n")
	}

	sb.WriteString("}\n")
	return sb.String()
}
