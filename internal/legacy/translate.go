package legacy

import (
	"sync"

	"pz-web-backend/internal/i18n"
)

// Legacy wrapper:
// 仅用于兼容旧的全局翻译访问方式（历史遗留）。
// 新代码请注入并使用 `internal/i18n.Loader`（例如通过 internal/application/i18napp）。
type TranslationMap = i18n.TranslationMap

var (
	defaultI18nMu sync.Mutex
	loaderByDir   = map[string]*i18n.Loader{}
)

// Deprecated: GetTranslationMapLegacy 仅用于兼容旧调用路径。
// 新代码优先通过 internal/application/i18napp。
func GetTranslationMapLegacy(baseGameDir string, lang string) TranslationMap {
	if baseGameDir == "" {
		return i18n.TranslationMap{}
	}
	return loaderFor(baseGameDir).GetTranslationMap(lang)
}

// Deprecated: TranslateKeyLegacy 仅用于兼容旧调用路径。
// 新代码请直接使用 internal/i18n.TranslateKey。
func TranslateKeyLegacy(t TranslationMap, key string, contextPrefix string) (label, tooltip string) {
	return i18n.TranslateKey(t, key, contextPrefix)
}

func loaderFor(baseGameDir string) *i18n.Loader {
	defaultI18nMu.Lock()
	defer defaultI18nMu.Unlock()

	if l, ok := loaderByDir[baseGameDir]; ok {
		return l
	}

	l := i18n.NewLoader(baseGameDir)
	loaderByDir[baseGameDir] = l
	return l
}
