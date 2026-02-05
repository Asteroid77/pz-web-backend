package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Clean(filepath.Join(".", "..", "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("repo root not found: %v", err)
	}
	return root
}

func TestRoutes_IndexOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	root := repoRoot(t)
	r := NewEngine(Config{
		BaseDataDir: filepath.Join(root, "testdata", "mock_zomboid"),
		BaseGameDir: filepath.Join(root, "testdata", "mock_media"),
		ServerName:  "",
		DevMode:     true,
		Build:       BuildInfo{Version: "test", GithubRepo: "test/repo"},
		ContentFS:   os.DirFS(root),
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if w.Body.Len() == 0 {
		t.Fatalf("expected non-empty body")
	}
}

func TestRoutes_I18nOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	root := repoRoot(t)
	r := NewEngine(Config{
		BaseDataDir: filepath.Join(root, "testdata", "mock_zomboid"),
		BaseGameDir: filepath.Join(root, "testdata", "mock_media"),
		ServerName:  "",
		DevMode:     true,
		Build:       BuildInfo{Version: "test", GithubRepo: "test/repo"},
		ContentFS:   os.DirFS(root),
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/i18n?lang=EN", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Languages []LanguageOption  `json:"languages"`
		UI        map[string]string `json:"ui"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v body=%s", err, w.Body.String())
	}
	if len(resp.Languages) == 0 {
		t.Fatalf("expected languages, got empty")
	}
	if resp.UI == nil || len(resp.UI) == 0 {
		t.Fatalf("expected ui resources")
	}
}

func TestRoutes_ConfigServerOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	root := repoRoot(t)

	r := NewEngine(Config{
		BaseDataDir: filepath.Join(root, "testdata", "mock_zomboid"),
		BaseGameDir: filepath.Join(root, "testdata", "mock_media"),
		ServerName:  "",
		DevMode:     true,
		Build:       BuildInfo{Version: "test", GithubRepo: "test/repo"},
		ContentFS:   os.DirFS(root),
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/config/server?lang=EN", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
