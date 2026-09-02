package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"et-interview/internal/db"
	"et-interview/internal/demo"
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
	return Wires{DB: d, Demo: demo.NewStore(d), WebRoot: t.TempDir()}
}

func TestDemoReturnsSeededValues(t *testing.T) {
	router := Routes(newTestWires(t))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/demo", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body)
	}

	var body struct {
		Values []float64 `json:"values"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode %s: %v", rec.Body, err)
	}

	want := []float64{1, 2, 3, 4, 5}
	if len(body.Values) != len(want) {
		t.Fatalf("got %v, want %v", body.Values, want)
	}
	for i := range want {
		if body.Values[i] != want[i] {
			t.Errorf("value %d: got %v, want %v", i, body.Values[i], want[i])
		}
	}
}

// A failing query must surface as a 500, not as a 200 carrying an empty list.
func TestDemoReportsStoreFailure(t *testing.T) {
	wires := newTestWires(t)
	router := Routes(wires)
	if err := wires.DB.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/demo", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want %d (body: %s)", rec.Code, http.StatusInternalServerError, rec.Body)
	}
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
	if ct := rec.Header().Get("Content-Type"); ct != "" && ct[:9] == "text/html" {
		t.Errorf("got HTML content type %q for an API path", ct)
	}
}
