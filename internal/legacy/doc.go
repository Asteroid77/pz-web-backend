// Package legacy 提供历史兼容 API（Legacy wrappers）。
//
// 该包存在的唯一目的：在重构期间保持旧的调用入口可用。
// 新代码请使用 internal/application/*（例如 modsapp/updateapp/configapp/i18napp），避免直接依赖 legacy。
package legacy
