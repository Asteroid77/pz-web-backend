package main

// ConfigOption 下拉菜单的单个选项
type ConfigOption struct {
	Value string `json:"value"` // 选项的值 (例如 "1", "2" 或 "true")
	Label string `json:"label"` // 选项的显示文本 (例如 "15 分钟")
}

// ConfigItem 单个配置项
type ConfigItem struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Label   string `json:"label"`
	Tooltip string `json:"tooltip"`
	Section string `json:"section"`
	// 如果这个字段不为空，前端应该渲染成 Select 框而不是 Input
	Options []ConfigOption `json:"options,omitempty"`
}
type LanguageOption struct {
	Code string `json:"code"`
	Name string `json:"name"`
}
type SaveRequest struct {
	Items   []ConfigItem `json:"items"` // 只需要 Key 和 Value
	Restart bool         `json:"restart"`
}
