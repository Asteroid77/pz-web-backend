package mods

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanLocalMods_FindsAndDedupesByModIDAndWorkshopID(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(root, "steamapps/workshop/content/108600/123456/mods/MyMod/42")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	modInfo := "name=MyMod\nid=mod.id\n"
	if err := os.WriteFile(filepath.Join(base, "mod.info"), []byte(modInfo), 0o644); err != nil {
		t.Fatalf("write mod.info: %v", err)
	}

	// 同一个 workshop 下再放一个同id的 mod.info，应覆盖/去重
	base2 := filepath.Join(root, "steamapps/workshop/content/108600/123456/mods/MyMod/mod.info")
	if err := os.MkdirAll(filepath.Dir(base2), 0o755); err != nil {
		t.Fatalf("mkdir2: %v", err)
	}
	modInfo2 := "name=MyModNew\nid=mod.id\ndescription=desc\n"
	if err := os.WriteFile(base2, []byte(modInfo2), 0o644); err != nil {
		t.Fatalf("write mod.info2: %v", err)
	}

	mods, err := ScanLocalMods(root)
	if err != nil {
		t.Fatalf("ScanLocalMods: %v", err)
	}
	if len(mods) != 1 {
		t.Fatalf("len(mods)=%d", len(mods))
	}
	if mods[0].WorkshopID != "123456" || mods[0].ModID != "mod.id" || mods[0].Name != "MyModNew" {
		t.Fatalf("unexpected mod: %+v", mods[0])
	}
}
