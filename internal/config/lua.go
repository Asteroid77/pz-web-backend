package config

import (
	"fmt"
	"strconv"
	"strings"
)

func isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func isBool(s string) bool {
	return s == "true" || s == "false"
}

func formatLuaValue(val string) string {
	val = strings.TrimSpace(val)
	if isBool(val) || isNumber(val) {
		return val
	}
	if !strings.HasPrefix(val, "\"") {
		return fmt.Sprintf("\"%s\"", val)
	}
	return val
}
