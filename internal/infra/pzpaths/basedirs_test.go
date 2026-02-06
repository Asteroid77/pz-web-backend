package pzpaths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveBaseDirs_DevMode_PicksFirstExisting(t *testing.T) {
	root := t.TempDir()

	wantData := filepath.Join(root, "testdata/mock_zomboid")
	wantGame := filepath.Join(root, "testdata/mock_media_full/mock_media")

	if err := os.MkdirAll(wantData, 0o755); err != nil {
		t.Fatalf("mkdir data: %v", err)
	}
	if err := os.MkdirAll(wantGame, 0o755); err != nil {
		t.Fatalf("mkdir game: %v", err)
	}

	got := ResolveBaseDirs(ResolveBaseDirsOptions{
		DevMode: true,
		CWD:     root,
	})

	if got.DataDir != wantData {
		t.Fatalf("DataDir=%q want %q", got.DataDir, wantData)
	}
	if got.GameDir != wantGame {
		t.Fatalf("GameDir=%q want %q", got.GameDir, wantGame)
	}
}

func TestResolveBaseDirs_DevMode_PrefersFullFixtureRoots(t *testing.T) {
	root := t.TempDir()

	wantData := filepath.Join(root, "testdata/mock_zomboid_full/mock_zomboid")
	wantGame := filepath.Join(root, "testdata/mock_media_full/mock_media")

	if err := os.MkdirAll(wantData, 0o755); err != nil {
		t.Fatalf("mkdir data: %v", err)
	}
	if err := os.MkdirAll(wantGame, 0o755); err != nil {
		t.Fatalf("mkdir game: %v", err)
	}

	got := ResolveBaseDirs(ResolveBaseDirsOptions{
		DevMode: true,
		CWD:     root,
	})

	if got.DataDir != wantData {
		t.Fatalf("DataDir=%q want %q", got.DataDir, wantData)
	}
	if got.GameDir != wantGame {
		t.Fatalf("GameDir=%q want %q", got.GameDir, wantGame)
	}
}

func TestResolveBaseDirs_ProdMode_UsesEnvOrDefaults(t *testing.T) {
	got := ResolveBaseDirs(ResolveBaseDirsOptions{
		DevMode:       false,
		EnvDataDir:    "",
		EnvInstallDir: "",
	})
	if got.DataDir != "/home/steam/Zomboid" {
		t.Fatalf("DataDir=%q", got.DataDir)
	}
	if got.GameDir != filepath.Join("/opt/pzserver", "media") {
		t.Fatalf("GameDir=%q", got.GameDir)
	}

	got = ResolveBaseDirs(ResolveBaseDirsOptions{
		DevMode:       false,
		EnvDataDir:    "/x/data",
		EnvInstallDir: "/x/install",
	})
	if got.DataDir != "/x/data" {
		t.Fatalf("DataDir=%q", got.DataDir)
	}
	if got.GameDir != filepath.Join("/x/install", "media") {
		t.Fatalf("GameDir=%q", got.GameDir)
	}
}
