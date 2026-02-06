package i18napp

import (
	"path/filepath"
	"sort"
	"strings"

	"pz-web-backend/internal/i18n"
	"pz-web-backend/internal/infra/fs"
)

type LanguageOption struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Service struct {
	BaseGameDir string
	FS          fs.FS
}

type Response struct {
	Lang      string            `json:"lang"`
	Languages []LanguageOption  `json:"languages"`
	UI        map[string]string `json:"ui"`
}

func (s Service) Get(lang string) Response {
	languages := s.listLanguages()
	active := resolveLang(lang, languages)

	uiResources, ok := i18n.WebUIResources[active]
	if !ok {
		uiResources = i18n.WebUIResources["EN"]
		active = "EN"
	}

	return Response{
		Lang:      active,
		Languages: languages,
		UI:        uiResources,
	}
}

func (s Service) ResolveLang(lang string) string {
	return resolveLang(lang, s.listLanguages())
}

func (s Service) listLanguages() []LanguageOption {
	if s.FS == nil {
		return defaultLanguages()
	}

	translateDir := filepath.Join(s.BaseGameDir, "lua/shared/Translate")
	files, err := s.FS.ReadDir(translateDir)
	if err != nil {
		return defaultLanguages()
	}

	var languages []LanguageOption
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		code := strings.ToUpper(f.Name())
		name, ok := i18n.LangNameMap[code]
		if !ok {
			name = code
		}
		languages = append(languages, LanguageOption{Code: code, Name: name})
	}

	if len(languages) == 0 {
		return defaultLanguages()
	}

	sort.Slice(languages, func(i, j int) bool {
		a := languages[i].Code
		b := languages[j].Code
		if a == b {
			return false
		}
		pa := langPriority(a)
		pb := langPriority(b)
		if pa != pb {
			return pa < pb
		}
		return a < b
	})
	return languages
}

func defaultLanguages() []LanguageOption {
	return []LanguageOption{
		{Code: "CN", Name: langNameOrCode("CN")},
		{Code: "EN", Name: langNameOrCode("EN")},
	}
}

func langNameOrCode(code string) string {
	if name, ok := i18n.LangNameMap[code]; ok {
		return name
	}
	return code
}

func langPriority(code string) int {
	switch code {
	case "CN":
		return 0
	case "EN":
		return 1
	default:
		return 2
	}
}

func resolveLang(requested string, languages []LanguageOption) string {
	req := strings.ToUpper(strings.TrimSpace(requested))
	if req == "" {
		req = "CN"
	}

	available := map[string]struct{}{}
	for _, l := range languages {
		available[l.Code] = struct{}{}
	}

	if _, ok := available[req]; ok {
		return req
	}
	if _, ok := available["EN"]; ok {
		return "EN"
	}
	if len(languages) > 0 {
		return languages[0].Code
	}
	return "EN"
}
