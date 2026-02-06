package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestHandleRestartPanel_ReturnsOKAndTriggersExit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	oldDelay := panelRestartDelay
	oldSleep := panelRestartSleep
	oldExit := panelRestartExit
	t.Cleanup(func() {
		panelRestartDelay = oldDelay
		panelRestartSleep = oldSleep
		panelRestartExit = oldExit
	})

	exitCh := make(chan int, 1)
	panelRestartDelay = 0
	panelRestartSleep = func(time.Duration) {}
	panelRestartExit = func(code int) { exitCh <- code }

	r := gin.New()
	r.POST("/api/service/restart", handleRestartPanel)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/service/restart", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if body := w.Body.String(); !strings.Contains(body, "Restarting") {
		t.Fatalf("unexpected body=%s", body)
	}

	select {
	case code := <-exitCh:
		if code != 0 {
			t.Fatalf("exit=%d", code)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected exit to be called")
	}
}
