package i18napp

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"pz-web-backend/internal/infra/fs"
)

type readDirErrFS struct {
	err error
}

func (readDirErrFS) MkdirAll(string, os.FileMode) error          { return nil }
func (readDirErrFS) WriteFile(string, []byte, os.FileMode) error { return nil }
func (e readDirErrFS) ReadDir(string) ([]os.DirEntry, error)     { return nil, e.err }

func TestService_Get_SortsLanguages_CN_EN_First(t *testing.T) {
	base := t.TempDir()
	translateDir := filepath.Join(base, "lua/shared/Translate")

	osfs := fs.OSFS{}
	for _, code := range []string{"DE", "EN", "CN", "ZZ"} {
		if err := osfs.MkdirAll(filepath.Join(translateDir, code), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}

	svc := Service{
		BaseGameDir: base,
		FS:          osfs,
	}
	resp := svc.Get("DE")

	if resp.Lang != "DE" {
		t.Fatalf("Lang=%q", resp.Lang)
	}

	var gotCodes []string
	for _, l := range resp.Languages {
		gotCodes = append(gotCodes, l.Code)
	}
	wantCodes := []string{"CN", "EN", "DE", "ZZ"}
	if !reflect.DeepEqual(gotCodes, wantCodes) {
		t.Fatalf("codes=%v want %v", gotCodes, wantCodes)
	}

	nameByCode := map[string]string{}
	for _, l := range resp.Languages {
		nameByCode[l.Code] = l.Name
	}
	if nameByCode["CN"] == "" || nameByCode["EN"] == "" || nameByCode["DE"] == "" {
		t.Fatalf("unexpected empty name: %+v", nameByCode)
	}
	if nameByCode["ZZ"] != "ZZ" {
		t.Fatalf("ZZ name=%q", nameByCode["ZZ"])
	}
}

func TestService_ResolveLang_FallbackToEN(t *testing.T) {
	base := t.TempDir()
	translateDir := filepath.Join(base, "lua/shared/Translate")

	osfs := fs.OSFS{}
	if err := osfs.MkdirAll(filepath.Join(translateDir, "EN"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	svc := Service{BaseGameDir: base, FS: osfs}
	if got := svc.ResolveLang("CN"); got != "EN" {
		t.Fatalf("got=%q", got)
	}
}

func TestService_Get_FallbackLanguagesWhenReadDirFails(t *testing.T) {
	svc := Service{
		BaseGameDir: "/does-not-matter",
		FS:          readDirErrFS{err: errors.New("boom")},
	}
	resp := svc.Get("")

	var gotCodes []string
	for _, l := range resp.Languages {
		gotCodes = append(gotCodes, l.Code)
	}
	wantCodes := []string{"CN", "EN"}
	if !reflect.DeepEqual(gotCodes, wantCodes) {
		t.Fatalf("codes=%v want %v", gotCodes, wantCodes)
	}
	if resp.Lang != "CN" {
		t.Fatalf("Lang=%q", resp.Lang)
	}
}
