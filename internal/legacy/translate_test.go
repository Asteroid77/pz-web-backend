package legacy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetTranslationMapLegacy_LoadsFromDisk(t *testing.T) {
	base := t.TempDir()
	uiDir := filepath.Join(base, "lua/shared/Translate/EN")
	if err := os.MkdirAll(uiDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	uiPath := filepath.Join(uiDir, "UI_EN.txt")
	content := `{
UI_ServerOption_Foo_tooltip = "Hello",
}`
	if err := os.WriteFile(uiPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	dict := GetTranslationMapLegacy(base, "EN")
	if got := dict["UI_ServerOption_Foo_tooltip"]; got != "Hello" {
		t.Fatalf("unexpected value: %q", got)
	}
}

func TestTranslateKeyLegacy_Passthrough(t *testing.T) {
	dict := TranslationMap{
		"UI_ServerOption_Foo":         "Foo Label",
		"UI_ServerOption_Foo_tooltip": "Foo Tip",
	}
	label, tip := TranslateKeyLegacy(dict, "Foo", "UI_ServerOption_")
	if label != "Foo Label" {
		t.Fatalf("label=%q", label)
	}
	if tip != "Foo Tip" {
		t.Fatalf("tip=%q", tip)
	}
}
