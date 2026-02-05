package updateapp

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	infrart "pz-web-backend/internal/infra/runtime"
	sysupdate "pz-web-backend/internal/system/update"
)

type fakeRuntime struct {
	executable string

	rename func(oldpath, newpath string) error
	chmod  func(name string, mode os.FileMode) error
	create func(name string) (*os.File, error)
	sleep  func(d time.Duration)
	exit   func(code int)
}

func (r fakeRuntime) Executable() (string, error) { return r.executable, nil }
func (r fakeRuntime) Rename(oldpath, newpath string) error {
	if r.rename != nil {
		return r.rename(oldpath, newpath)
	}
	return os.Rename(oldpath, newpath)
}
func (r fakeRuntime) Chmod(name string, mode os.FileMode) error {
	if r.chmod != nil {
		return r.chmod(name, mode)
	}
	return os.Chmod(name, mode)
}
func (r fakeRuntime) Create(name string) (*os.File, error) {
	if r.create != nil {
		return r.create(name)
	}
	return os.Create(name)
}
func (r fakeRuntime) Sleep(d time.Duration) {
	if r.sleep != nil {
		r.sleep(d)
		return
	}
	time.Sleep(d)
}
func (r fakeRuntime) Exit(code int) {
	if r.exit != nil {
		r.exit(code)
		return
	}
}

func TestService_PerformUpdate_DevMode_NoReplace(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bin"))
	}))
	defer srv.Close()

	tmp := t.TempDir()
	tmpPath := filepath.Join(tmp, "newbin")

	svc := NewService(true, sysupdate.Service{})
	svc.HTTPClient = srv.Client()
	svc.TmpPath = tmpPath
	svc.Runtime = fakeRuntime{
		executable: filepath.Join(tmp, "oldbin"),
		rename: func(oldpath, newpath string) error {
			t.Fatalf("rename should not be called in dev mode")
			return nil
		},
		sleep: func(d time.Duration) {},
		exit: func(code int) {
			t.Fatalf("exit should not be called in dev mode")
		},
	}

	if err := svc.PerformUpdate(srv.URL); err != nil {
		t.Fatalf("err=%v", err)
	}
	if _, err := os.Stat(tmpPath); err != nil {
		t.Fatalf("expected tmp written: %v", err)
	}
}

func TestService_PerformUpdate_ReplacesBinaryAndExit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("bin"))
	}))
	defer srv.Close()

	tmp := t.TempDir()
	oldPath := filepath.Join(tmp, "oldbin")
	if err := os.WriteFile(oldPath, []byte("old"), 0o755); err != nil {
		t.Fatalf("write old: %v", err)
	}
	tmpPath := filepath.Join(tmp, "newbin")

	done := make(chan struct{})
	svc := NewService(false, sysupdate.Service{})
	svc.HTTPClient = srv.Client()
	svc.TmpPath = tmpPath
	svc.Runtime = fakeRuntime{
		executable: oldPath,
		sleep:      func(d time.Duration) {},
		exit:       func(code int) { close(done) },
	}

	if err := svc.PerformUpdate(srv.URL); err != nil {
		t.Fatalf("err=%v", err)
	}

	got, err := os.ReadFile(oldPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !bytes.Equal(got, []byte("bin")) {
		t.Fatalf("got=%q", string(got))
	}

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("expected exit to be called")
	}
}

func TestService_NewService_DefaultRuntime(t *testing.T) {
	svc := NewService(true, sysupdate.Service{})
	if svc.Runtime == nil {
		t.Fatalf("expected runtime not nil")
	}
	// compile-time check
	var _ infrart.Ops = svc.Runtime
}
