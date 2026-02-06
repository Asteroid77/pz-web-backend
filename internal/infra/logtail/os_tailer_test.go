package logtail

import (
	"context"
	"testing"
)

func TestOSTailer_Tail_RequiresPath(t *testing.T) {
	tailer := OSTailer{}
	rc, err := tailer.Tail(context.Background(), "", 10)
	if err == nil {
		t.Fatalf("expected error")
	}
	if rc != nil {
		t.Fatalf("expected nil rc")
	}
}
