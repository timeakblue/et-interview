package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"et-interview/internal/db"
)

func newTestWires(t *testing.T) Wires {
	t.Helper()
	gin.SetMode(gin.TestMode)

	d, err := db.Connect(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })

	if err := db.Migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return Wires{DB: d, WebRoot: t.TempDir()}
}

// Unmatched /api/ paths must 404 rather than falling through to the static
// file server, which would answer an API call with HTML.
func TestUnknownAPIPathIsNotFound(t *testing.T) {
	router := Routes(newTestWires(t))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/nope", nil))

	if rec.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusNotFound)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "" && len(ct) >= 9 && ct[:9] == "text/html" {
		t.Errorf("got HTML content type %q for an API path", ct)
	}
}
