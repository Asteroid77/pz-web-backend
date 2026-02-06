package mods

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type memStore struct {
	m map[string]ModInfo
}

func (s *memStore) Load() (map[string]ModInfo, error) { return s.m, nil }
func (s *memStore) Save(m map[string]ModInfo) error   { s.m = m; return nil }

func TestExtractModID_MultipleAndDedup(t *testing.T) {
	desc := "Hello\nMod ID: A\nMod ID: A\nMod ID: B<br>xxx"
	got := extractModID(desc)
	if got != "A,B" {
		t.Fatalf("got=%q", got)
	}
}

func TestWorkshopClient_FetchAndCache(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		id := r.Form.Get("publishedfileids[0]")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"response":{"resultcount":1,"publishedfiledetails":[{"publishedfileid":"` + id + `","title":"T","description":"Mod ID: X"}]}}`))
	}))
	defer srv.Close()

	httpClient := &http.Client{Timeout: 2 * time.Second}
	store := &memStore{m: map[string]ModInfo{}}
	c, err := NewWorkshopClient(httpClient, srv.URL, store)
	if err != nil {
		t.Fatalf("NewWorkshopClient: %v", err)
	}

	info, err := c.FetchWorkshopInfo("999")
	if err != nil {
		t.Fatalf("FetchWorkshopInfo: %v", err)
	}
	if info.WorkshopID != "999" || info.Name != "T" || info.ModID != "X" {
		t.Fatalf("unexpected info: %+v", info)
	}

	// 第二次应命中内存缓存（不验证请求次数，至少应一致）
	info2, err := c.FetchWorkshopInfo("999")
	if err != nil {
		t.Fatalf("FetchWorkshopInfo2: %v", err)
	}
	if info2 != info {
		t.Fatalf("cache mismatch: %+v vs %+v", info2, info)
	}

	// flush 为异步；这里不强依赖，但至少在内存 store 中最终可见（best-effort）
	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		if v, ok := store.m["999"]; ok && v.ModID == "X" {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("expected store saved, got=%v", store.m)
}

func TestWorkshopClient_ModNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"response":{"resultcount":0,"publishedfiledetails":[]}}`))
	}))
	defer srv.Close()

	c, err := NewWorkshopClient(http.DefaultClient, srv.URL, nil)
	if err != nil {
		t.Fatalf("NewWorkshopClient: %v", err)
	}
	_, err = c.FetchWorkshopInfo("1")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("err=%v", err)
	}
}
