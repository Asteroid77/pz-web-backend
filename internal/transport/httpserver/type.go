package httpserver

import (
	"pz-web-backend/internal/config"
	"pz-web-backend/internal/mods"
)

type ConfigOption = config.Option
type ConfigItem = config.Item
type ModInfo = mods.ModInfo

type LanguageOption struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type SaveRequest struct {
	Items   []ConfigItem `json:"items"`
	Restart bool         `json:"restart"`
}
