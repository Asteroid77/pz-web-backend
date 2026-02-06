package config

// Option 下拉菜单的单个选项。
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// Item 单个配置项。
type Item struct {
	Key     string   `json:"key"`
	Value   string   `json:"value"`
	Label   string   `json:"label"`
	Tooltip string   `json:"tooltip"`
	Section string   `json:"section"`
	Options []Option `json:"options,omitempty"`
}
