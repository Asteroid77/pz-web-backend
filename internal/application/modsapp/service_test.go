package modsapp

import (
	"os"
	"path/filepath"
	"testing"

	"pz-web-backend/internal/mods"
)

type stubWorkshop struct {
	m   map[string]mods.ModInfo
	err error
}

func (s stubWorkshop) FetchWorkshopInfo(workshopID string) (mods.ModInfo, error) {
	if s.err != nil {
		return mods.ModInfo{}, s.err
	}
	if v, ok := s.m[workshopID]; ok {
		return v, nil
	}
	return mods.ModInfo{}, os.ErrNotExist
}

func TestService_ListLocalMods(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, "steamapps/workshop/content/108600/123/mods/A")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(base, "mod.info"), []byte("id=a\nname=A\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	svc := Service{InstallDir: root}
	items, err := svc.ListLocalMods()
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if len(items) != 1 || items[0].WorkshopID != "123" || items[0].ModID != "a" {
		t.Fatalf("items=%+v", items)
	}
}

func TestService_Lookup_PrefersLocal(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, "steamapps/workshop/content/108600/999/mods/A")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(base, "mod.info"), []byte("id=a\nname=A\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	svc := Service{
		InstallDir: root,
		Workshop: stubWorkshop{
			m: map[string]mods.ModInfo{
				"999": {WorkshopID: "999", ModID: "x", Name: "X"},
			},
		},
	}
	res, _ := svc.Lookup([]string{"999"})
	if len(res) != 1 || res[0].Source != "local" || len(res[0].Mods) != 1 || res[0].Mods[0].ModID != "a" {
		t.Fatalf("res=%+v", res)
	}
}

func TestService_Lookup_FallsBackToSteam(t *testing.T) {
	svc := Service{
		InstallDir: t.TempDir(),
		Workshop: stubWorkshop{
			m: map[string]mods.ModInfo{
				"1": {WorkshopID: "1", ModID: "m", Name: "N"},
			},
		},
	}
	res, _ := svc.Lookup([]string{"1"})
	if len(res) != 1 || res[0].Source != "steam" || len(res[0].Mods) != 1 || res[0].Mods[0].ModID != "m" {
		t.Fatalf("res=%+v", res)
	}
}
