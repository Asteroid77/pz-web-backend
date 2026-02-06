package httpserver

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type staticTailer struct {
	rc io.ReadCloser
}

func (t staticTailer) Tail(ctx context.Context, path string, lines int) (io.ReadCloser, error) {
	return t.rc, nil
}

func TestHandleStreamLogs_LongLinesNotTruncated(t *testing.T) {
	gin.SetMode(gin.TestMode)

	longLine := strings.Repeat("a", 70*1024)
	tailer := staticTailer{rc: io.NopCloser(strings.NewReader(longLine + "\n"))}

	app := App{
		LogPath:   "/does-not-matter",
		LogTailer: tailer,
	}

	r := gin.New()
	r.GET("/api/logs/stream", app.handleStreamLogs)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/logs/stream", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body_len=%d", w.Code, w.Body.Len())
	}

	body := w.Body.String()
	if strings.Contains(body, "event: error") {
		t.Fatalf("unexpected error event")
	}
	if !strings.Contains(body, "data: ") {
		t.Fatalf("expected data event, body_len=%d", len(body))
	}

	// Verify both ends of the payload are present (avoids matching a 70KB substring).
	if !strings.Contains(body, longLine[:1024]) {
		t.Fatalf("missing prefix payload")
	}
	if !strings.Contains(body, longLine[len(longLine)-1024:]) {
		t.Fatalf("missing suffix payload")
	}
}
