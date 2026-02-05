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
	translateDir := filepath.Join(s.BaseGameDir, "lua/shared/Translate")
	files, _ := s.FS.ReadDir(translateDir)

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
	return languages
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
