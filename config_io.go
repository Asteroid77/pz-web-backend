package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
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

	// 尝试回退到 CN (假设 CN 最全) 或者 EN
	fallbackLang := "CN"
	if lang == "CN" {
		fallbackLang = "EN"
	}

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
func getSandboxSection(key string, language string) string {
	// 僵尸特性 (Zombie Lore) - 这些 Key 名字很短且通用，精确匹配
	switch key {
	case "Speed", "Strength", "Toughness", "Transmission", "Mortality",
		"Reanimate", "Cognition", "Memory", "Decomp", "Sight", "Hearing",
		"Smell", "ThumpNoChasing", "ThumpOnConstruction", "ActiveOnly", "TriggerHouseAlarm":
		return getSafeSectionName(language, SecZombieLore)
	}

	// 模式匹配
	// 转换为小写方便匹配
	k := strings.ToLower(key)

	// 物资 (Loot) - 通常以 Loot 结尾
	if strings.HasSuffix(k, "loot") {
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

func ParseServerTestIni(path string, lang string) ([]ConfigItem, error) {
	dict := GetTranslationMap(lang)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var items []ConfigItem
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 直接跳过注释
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			// ServerOption 翻译
			_, tooltip := TranslateKey(dict, key, "UI_ServerOption_")
			section := getServerOptionSectionName(key, lang)

			items = append(items, ConfigItem{
				Key:     key,
				Value:   val,
				Label:   tooltip,
				Section: section,
			})
		}
	}
	return items, nil
}

// ParseSandboxLua 解析 SandboxVars.lua
func ParseSandboxLua(path string, lang string) ([]ConfigItem, error) {
	dict := GetTranslationMap(lang)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var items []ConfigItem
	re := regexp.MustCompile(`^\s*([A-Za-z0-9_]+)\s*=\s*(.*),`)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {

		line := strings.TrimSpace(scanner.Text())

		// 直接跳过注释
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			key := matches[1]
			val := matches[2]
			// 去掉 Lua 字符串的引号 (如果有)
			val = strings.Trim(val, "\"")

			// 获取基础 Label 和 Tooltip
			label, tooltip := TranslateKey(dict, key, "Sandbox_")

			// 尝试获取下拉选项
			options := GetSandboxOptions(dict, key)

			// 获取分组
			section := getSandboxSection(key, lang)

			items = append(items, ConfigItem{
				Key:     key,
				Value:   val,
				Label:   label,
				Tooltip: tooltip,
				Options: options, // 如果没选项，这里就是 nil/empty，前端显示 Input
				Section: section,
			})
		}
	}
	return items, nil
}

func GenerateServerTestIni(items []ConfigItem) string {
	var sb strings.Builder

	// 为了美观，我们可以稍微分个组，或者直接按顺序写入
	for _, item := range items {
		// INI 格式: Key=Value
		sb.WriteString(fmt.Sprintf("%s=%s\n", item.Key, item.Value))
	}

	return sb.String()
}
func GenerateSandboxLua(items []ConfigItem) string {
	var sb strings.Builder

	sb.WriteString("SandboxVars = {\n")

	for _, item := range items {
		val := item.Value
		// Lua 的值处理比较微妙：数字和布尔值不应该加引号，但字符串必须加
		// 简单策略：全部当字符串处理，僵毁会自己强转
		sb.WriteString(fmt.Sprintf("    %s = \"%s\",\n", item.Key, val))
	}
	sb.WriteString("}\n")
	return sb.String()
}
