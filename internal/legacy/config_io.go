package legacy

import (
	"fmt"

	"pz-web-backend/internal/config"
	"pz-web-backend/internal/i18n"
)

// 注意：旧实现已迁移到 internal/config，并通过本文件提供兼容层。

// Deprecated: ParseServerTestIniLegacy 仅用于兼容旧调用路径。
// 新代码优先通过 internal/application/configapp 或直接使用 internal/config.Service。
func ParseServerTestIniLegacy(baseGameDir string, path string, lang string) ([]config.Item, error) {
	if baseGameDir == "" {
		return nil, fmt.Errorf("baseGameDir is required")
	}
	return defaultConfigService(baseGameDir).ParseServerINI(path, lang)
}

// Deprecated: ParseSandboxLuaLegacy 仅用于兼容旧调用路径。
func ParseSandboxLuaLegacy(baseGameDir string, path string, lang string) ([]config.Item, error) {
	if baseGameDir == "" {
		return nil, fmt.Errorf("baseGameDir is required")
	}
	return defaultConfigService(baseGameDir).ParseSandboxLua(path, lang)
}

// Deprecated: GenerateServerTestIniLegacy 仅用于兼容旧调用路径。
func GenerateServerTestIniLegacy(items []config.Item) string {
	return config.Service{}.GenerateServerINI(items)
}

// Deprecated: GenerateSandboxLuaLegacy 仅用于兼容旧调用路径。
func GenerateSandboxLuaLegacy(items []config.Item) string {
	return config.Service{}.GenerateSandboxLua(items)
}

func defaultConfigService(baseGameDir string) config.Service {
	return config.Service{
		I18n: loaderFor(baseGameDir),
		SectionLabel: func(lang string, sectionKey string) string {
			if langMap, ok := i18n.WebUIResources[lang]; ok {
				if val, ok := langMap[sectionKey]; ok && val != "" {
					return val
				}
			}
			if langMap, ok := i18n.WebUIResources["EN"]; ok {
				if val, ok := langMap[sectionKey]; ok && val != "" {
					return val
				}
			}
			return sectionKey
		},
	}
}
