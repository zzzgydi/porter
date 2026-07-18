package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConsoleHandler(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("console"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte("asset"), 0o644); err != nil {
		t.Fatal(err)
	}

	h := consoleHandler(dir)
	if h == nil {
		t.Fatal("expected handler")
	}

	for _, tc := range []struct {
		name       string
		method     string
		path       string
		status     int
		body       string
		cacheValue string
	}{
		{"spa route", http.MethodGet, "/projects/demo", http.StatusOK, "console", "no-cache"},
		{"asset", http.MethodGet, "/assets/app.js", http.StatusOK, "asset", "public, max-age=31536000, immutable"},
		{"unknown API", http.MethodGet, "/api/unknown", http.StatusNotFound, "", ""},
		{"SPA post", http.MethodPost, "/dashboard", http.StatusNotFound, "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
			if tc.body != "" && !strings.Contains(rec.Body.String(), tc.body) {
				t.Fatalf("body = %q, want it to contain %q", rec.Body.String(), tc.body)
			}
			if got := rec.Header().Get("Cache-Control"); got != tc.cacheValue {
				t.Fatalf("Cache-Control = %q, want %q", got, tc.cacheValue)
			}
		})
	}
}
