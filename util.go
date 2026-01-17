package main

import (
	"fmt"
	"strconv"
	"strings"
)

// 辅助函数：判断字符串是否为数字
func isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// 辅助函数：判断字符串是否为布尔值
func isBool(s string) bool {
	return s == "true" || s == "false"
}

// 辅助函数：格式化 Lua 值（决定是否加引号）
func formatLuaValue(val string) string {
	val = strings.TrimSpace(val)
	if isBool(val) || isNumber(val) {
		return val
	}
	// 如果本身没有引号，且不是数字/布尔，则加上引号
	if !strings.HasPrefix(val, "\"") {
		return fmt.Sprintf("\"%s\"", val)
	}
	return val
}
