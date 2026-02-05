package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pz-web-backend/internal/i18n"
)

func TestService_ParseServerINI_SectionAndTooltip(t *testing.T) {
	base := t.TempDir()
	langDir := filepath.Join(base, "lua/shared/Translate/EN")
	if err := os.MkdirAll(langDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	uiPath := filepath.Join(langDir, "UI_EN.txt")
	uiContent := `{
UI_ServerOption_SteamVAC_tooltip = "Steam VAC Tooltip",
}`
	if err := os.WriteFile(uiPath, []byte(uiContent), 0o644); err != nil {
		t.Fatalf("write ui: %v", err)
	}

	iniPath := filepath.Join(t.TempDir(), "servertest.ini")
	iniContent := "SteamVAC=true\nMap=Muldraugh, KY\n"
	if err := os.WriteFile(iniPath, []byte(iniContent), 0o644); err != nil {
		t.Fatalf("write ini: %v", err)
	}

	svc := Service{
		I18n: i18n.NewLoader(base),
		SectionLabel: func(lang, sectionKey string) string {
			return sectionKey
		},
	}

	items, err := svc.ParseServerINI(iniPath, "EN")
	if err != nil {
		t.Fatalf("ParseServerINI: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items)=%d", len(items))
	}

	if items[0].Key != "SteamVAC" || items[0].Label != "Steam VAC Tooltip" {
		t.Fatalf("unexpected item[0]=%+v", items[0])
	}
	if items[0].Section != SecAntiCheat {
		t.Fatalf("section=%q", items[0].Section)
	}
	if items[1].Section != SecMap {
		t.Fatalf("section=%q", items[1].Section)
	}
}

func TestService_ParseSandboxLua_NestedKeyOptionsAndSection(t *testing.T) {
	base := t.TempDir()
	langDir := filepath.Join(base, "lua/shared/Translate/EN")
	if err := os.MkdirAll(langDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	sandboxPath := filepath.Join(langDir, "Sandbox_EN.txt")
	sandboxContent := `{
Sandbox_Speed = "Speed",
Sandbox_Speed_tooltip = "Speed Tip",
Sandbox_DayLength = "DayLength",
Sandbox_DayLength_option1 = "Opt1",
Sandbox_DayLength_option2 = "Opt2",
}`
	if err := os.WriteFile(sandboxPath, []byte(sandboxContent), 0o644); err != nil {
		t.Fatalf("write sandbox: %v", err)
	}

	luaPath := filepath.Join(t.TempDir(), "SandboxVars.lua")
	luaContent := `SandboxVars = {
    VERSION = 6,
    DayLength = 1,
    ZombieLore = {
        Speed = 2,
    },
}`
	if err := os.WriteFile(luaPath, []byte(luaContent), 0o644); err != nil {
		t.Fatalf("write lua: %v", err)
	}

	svc := Service{
		I18n: i18n.NewLoader(base),
		SectionLabel: func(lang, sectionKey string) string {
			return sectionKey
		},
	}

	items, err := svc.ParseSandboxLua(luaPath, "EN")
	if err != nil {
		t.Fatalf("ParseSandboxLua: %v", err)
	}

	var dayLength *Item
	var speed *Item
	for i := range items {
		if items[i].Key == "DayLength" {
			dayLength = &items[i]
		}
		if items[i].Key == "ZombieLore.Speed" {
			speed = &items[i]
		}
	}
	if dayLength == nil || speed == nil {
		t.Fatalf("missing keys: dayLength=%v speed=%v", dayLength != nil, speed != nil)
	}

	if len(dayLength.Options) != 2 || dayLength.Options[0].Label != "Opt1" {
		t.Fatalf("unexpected options: %+v", dayLength.Options)
	}
	if speed.Section != SecZombieLore {
		t.Fatalf("speed section=%q", speed.Section)
	}
	if speed.Tooltip != "Speed Tip" || speed.Label != "Speed" {
		t.Fatalf("speed label/tooltip=%q/%q", speed.Label, speed.Tooltip)
	}
}

func TestService_GenerateSandboxLua_QuotesStringValues(t *testing.T) {
	svc := Service{}
	out := svc.GenerateSandboxLua([]Item{
		{Key: "Foo", Value: "bar"},
		{Key: "Zombies", Value: "3"},
	})

	if want := "Foo = \"bar\""; !strings.Contains(out, want) {
		t.Fatalf("expected line %q in output:\n%s", want, out)
	}
	if want := "Zombies = 3"; !strings.Contains(out, want) {
		t.Fatalf("expected line %q in output:\n%s", want, out)
	}
}
