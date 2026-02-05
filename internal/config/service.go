package config

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"pz-web-backend/internal/i18n"
)

// SectionLabelResolver 将 sectionKey 转换为展示用名称（可按语言本地化）。
// 例如：("CN","mods_workshop") -> "模组与工坊"
type SectionLabelResolver func(lang string, sectionKey string) string

type Service struct {
	I18n         *i18n.Loader
	SectionLabel SectionLabelResolver
}

func (s Service) ParseServerINI(path string, lang string) ([]Item, error) {
	dict := s.I18n.GetTranslationMap(lang)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var items []Item
	scanner := bufio.NewScanner(f)

	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		_, tooltip := i18n.TranslateKey(dict, key, "UI_ServerOption_")
		if tooltip == "" {
			tooltip = key
		}

		sectionKey := inferServerSectionKey(key)
		section := s.resolveSectionLabel(lang, sectionKey)

		items = append(items, Item{
			Key:     key,
			Value:   val,
			Label:   tooltip,
			Section: section,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s Service) ParseSandboxLua(path string, lang string) ([]Item, error) {
	dict := s.I18n.GetTranslationMap(lang)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var items []Item

	reTableStart := regexp.MustCompile(`^\s*([A-Za-z0-9_]+)\s*=\s*\{\s*$`)
	reTableEnd := regexp.MustCompile(`^\s*\},?\s*$`)
	reValue := regexp.MustCompile(`^\s*([A-Za-z0-9_]+)\s*=\s*(.*),`)

	scanner := bufio.NewScanner(f)
	currentContext := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "--") {
			continue
		}

		if matches := reTableStart.FindStringSubmatch(line); len(matches) == 2 {
			tableName := matches[1]
			if tableName != "SandboxVars" {
				currentContext = tableName
				continue
			}
		}

		if reTableEnd.MatchString(line) {
			currentContext = ""
			continue
		}

		if matches := reValue.FindStringSubmatch(line); len(matches) == 3 {
			rawKey := matches[1]
			val := matches[2]

			val = strings.TrimSuffix(val, ",")
			val = strings.Trim(val, "\"")

			fullKey := rawKey
			if currentContext != "" {
				fullKey = currentContext + "." + rawKey
			}

			label, tooltip := i18n.TranslateKey(dict, rawKey, "Sandbox_")
			options := sandboxOptions(dict, rawKey)

			sectionKey := inferSandboxSectionKey(fullKey)
			section := s.resolveSectionLabel(lang, sectionKey)

			items = append(items, Item{
				Key:     fullKey,
				Value:   val,
				Label:   label,
				Tooltip: tooltip,
				Options: options,
				Section: section,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s Service) GenerateServerINI(items []Item) string {
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("%s=%s\n", item.Key, item.Value))
	}
	return sb.String()
}

func (s Service) GenerateSandboxLua(items []Item) string {
	var sb strings.Builder

	globals := []Item{}
	nestedMap := make(map[string][]Item)

	for _, item := range items {
		if strings.Contains(item.Key, ".") {
			parts := strings.SplitN(item.Key, ".", 2)
			parent := parts[0]
			childKey := parts[1]

			subItem := item
			subItem.Key = childKey
			nestedMap[parent] = append(nestedMap[parent], subItem)
		} else {
			globals = append(globals, item)
		}
	}

	sb.WriteString("SandboxVars = {\n")
	sb.WriteString("    VERSION = 6,\n")

	for _, item := range globals {
		if item.Key == "VERSION" {
			continue
		}
		val := formatLuaValue(item.Value)
		sb.WriteString(fmt.Sprintf("    %s = %s,\n", item.Key, val))
	}

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

func (s Service) resolveSectionLabel(lang string, sectionKey string) string {
	if s.SectionLabel == nil {
		return sectionKey
	}
	label := s.SectionLabel(lang, sectionKey)
	if label == "" {
		return sectionKey
	}
	return label
}

func sandboxOptions(dict i18n.TranslationMap, key string) []Option {
	var options []Option
	prefix := "Sandbox_"
	for i := 1; ; i++ {
		optionKey := fmt.Sprintf("%s%s_option%d", prefix, key, i)
		if label, ok := dict[optionKey]; ok {
			options = append(options, Option{
				Value: fmt.Sprintf("%d", i),
				Label: label,
			})
		} else {
			break
		}
	}
	return options
}

// ---- section inference（返回key而不是展示文本）----

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

func inferServerSectionKey(key string) string {
	k := strings.ToLower(key)

	if strings.HasPrefix(k, "anticheat") || strings.Contains(k, "hack") || strings.Contains(k, "checksum") {
		return SecAntiCheat
	}
	if strings.HasPrefix(k, "map") || strings.HasPrefix(k, "spawn") {
		return SecMap
	}
	if strings.HasPrefix(k, "steam") || strings.HasPrefix(k, "workshop") || strings.HasPrefix(k, "mod") {
		if k == "steamvac" {
			return SecAntiCheat
		}
		if strings.Contains(k, "scoreboard") {
			return SecGeneral
		}
		return SecMods
	}
	if strings.HasPrefix(k, "discord") {
		return SecDiscord
	}
	if strings.Contains(k, "player") || strings.Contains(k, "pvp") || strings.Contains(k, "safehouse") {
		return SecPlayers
	}
	if strings.Contains(k, "client") || strings.Contains(k, "ping") || strings.Contains(k, "speed") {
		return SecClient
	}
	return SecGeneral
}

func inferSandboxSectionKey(fullKey string) string {
	parts := strings.Split(fullKey, ".")
	parent := ""
	key := fullKey
	if len(parts) > 1 {
		parent = parts[0]
		key = parts[1]
	}

	switch parent {
	case "ZombieLore":
		return SecZombieLore
	case "Map", "Basement":
		return SecMap
	case "MultiplierConfig":
		return SecCharXP
	case "ZombieConfig":
		return SecZombieSpawn
	}

	k := strings.ToLower(key)
	if strings.HasSuffix(k, "loot") || strings.Contains(k, "lootnew") {
		return SecLoot
	}
	if strings.Contains(k, "start") || strings.Contains(k, "time") || key == "DayLength" {
		return SecTime
	}
	if strings.Contains(k, "zombie") || strings.Contains(k, "respawn") ||
		strings.Contains(k, "pop") || strings.Contains(k, "group") ||
		strings.Contains(k, "redistribute") || key == "Distribution" {
		return SecZombieSpawn
	}
	if strings.Contains(k, "water") || strings.Contains(k, "elec") ||
		strings.Contains(k, "fire") || strings.Contains(k, "rot") ||
		strings.Contains(k, "alarm") || strings.Contains(k, "house") ||
		strings.Contains(k, "generator") || strings.Contains(k, "fuel") {
		return SecWorld
	}
	if strings.Contains(k, "car") || strings.Contains(k, "traffic") ||
		strings.Contains(k, "gas") || strings.Contains(k, "vehicle") {
		return SecVehicles
	}
	if strings.Contains(k, "xp") || strings.Contains(k, "stat") ||
		strings.Contains(k, "regen") || strings.Contains(k, "clothing") ||
		strings.Contains(k, "injury") || strings.Contains(k, "move") ||
		strings.Contains(k, "stamina") || strings.Contains(k, "nutrition") {
		return SecCharXP
	}
	if strings.Contains(k, "erosion") || strings.Contains(k, "farming") ||
		strings.Contains(k, "nature") || strings.Contains(k, "rain") ||
		strings.Contains(k, "temperature") || strings.Contains(k, "plant") ||
		strings.Contains(k, "compost") {
		return SecNature
	}
	if strings.Contains(k, "map") || strings.Contains(k, "chunk") {
		return SecMap
	}
	return SecGeneral
}
