package demo_test

import (
	"context"
	"path/filepath"
	"testing"

	"et-interview/internal/db"
	"et-interview/internal/demo"
)

// newTestDB opens a throwaway file-backed database with the schema applied.
// A temp file rather than ":memory:" so the read and write pools are exercised
// the same way they are in a running server.
func newTestDB(t *testing.T) db.DB {
	t.Helper()

	conn := filepath.Join(t.TempDir(), "test.db")
	d, err := db.Connect(context.Background(), conn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() {
		if err := d.Close(); err != nil {
			t.Errorf("close: %v", err)
		}
	})

	if err := db.Migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return d
}

func TestStoreValues(t *testing.T) {
	store := demo.NewStore(newTestDB(t))

	got, err := store.Values(context.Background())
	if err != nil {
		t.Fatalf("Values: %v", err)
	}

	want := []float64{1, 2, 3, 4, 5}
	if len(got) != len(want) {
		t.Fatalf("got %d values %v, want %d %v", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("value %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

// Migrate runs on every startup, so it has to be safe to run against a
// database that already has the schema.
func TestMigrateIsIdempotent(t *testing.T) {
	d := newTestDB(t)

	if err := db.Migrate(context.Background(), d); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}

	got, err := demo.NewStore(d).Values(context.Background())
	if err != nil {
		t.Fatalf("Values: %v", err)
	}
	if len(got) != 5 {
		t.Errorf("got %d values after re-running Migrate, want 5", len(got))
	}
}
