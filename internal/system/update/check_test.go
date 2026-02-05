package update

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestService_CheckUpdate_FindsAsset(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
  "tag_name":"v1.2.0",
  "assets":[
    {"name":"pz-web-backend_linux_amd64","browser_download_url":"https://example.com/bin"},
    {"name":"other","browser_download_url":"https://example.com/other"}
  ]
}`))
	}))
	defer srv.Close()

	svc := Service{
		HTTPClient:     srv.Client(),
		APIBase:        srv.URL,
		GithubRepo:     "x/y",
		CurrentVersion: "v1.0.0",
		OS:             "linux",
		Arch:           "amd64",
	}

	tag, url, err := svc.CheckUpdate()
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if tag != "v1.2.0" || url != "https://example.com/bin" {
		t.Fatalf("tag/url=%q/%q", tag, url)
	}
}

func TestService_CheckUpdate_DevSkips(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"v9.9.9","assets":[]}`))
	}))
	defer srv.Close()

	svc := Service{
		HTTPClient:     srv.Client(),
		APIBase:        srv.URL,
		GithubRepo:     "x/y",
		CurrentVersion: "dev",
	}

	tag, url, err := svc.CheckUpdate()
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if tag != "" || url != "" {
		t.Fatalf("expected empty, got %q/%q", tag, url)
	}
}
