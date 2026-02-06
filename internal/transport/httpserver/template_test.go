package httpserver

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestTemplates_ParseAndExecuteIndex(t *testing.T) {
	root := filepath.Clean(filepath.Join(".", "..", "..", ".."))
	tmpl := mustParseTemplates(os.DirFS(root))

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "index.html", map[string]any{"title": "test"}); err != nil {
		t.Fatalf("ExecuteTemplate(index.html): %v", err)
	}
	if buf.Len() == 0 {
		t.Fatalf("expected non-empty output")
	}
}
