package configapp

import (
	"path/filepath"
	"testing"

	"pz-web-backend/internal/infra/fs"
)

func TestResolveServerName_ExplicitOverrides(t *testing.T) {
	osfs := fs.OSFS{}
	got := ResolveServerName(osfs, "/does-not-matter", "myserver.ini")
	if got != "myserver" {
		t.Fatalf("got=%q", got)
	}
}

func TestResolveServerName_PrefersServertestIfPresent(t *testing.T) {
	base := t.TempDir()
	serverDir := filepath.Join(base, "Server")
	osfs := fs.OSFS{}
	if err := osfs.MkdirAll(serverDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := osfs.WriteFile(filepath.Join(serverDir, "servertest.ini"), []byte("SteamVAC=true\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := osfs.WriteFile(filepath.Join(serverDir, "abc.ini"), []byte("SteamVAC=true\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	got := ResolveServerName(osfs, base, "")
	if got != "servertest" {
		t.Fatalf("got=%q", got)
	}
}

func TestResolveServerName_SingleIni(t *testing.T) {
	base := t.TempDir()
	serverDir := filepath.Join(base, "Server")
	osfs := fs.OSFS{}
	if err := osfs.MkdirAll(serverDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := osfs.WriteFile(filepath.Join(serverDir, "myworld.ini"), []byte("SteamVAC=true\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	got := ResolveServerName(osfs, base, "")
	if got != "myworld" {
		t.Fatalf("got=%q", got)
	}
}
