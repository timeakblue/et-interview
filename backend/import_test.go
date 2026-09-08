package main

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunImportWithoutPathsFails(t *testing.T) {
	if code := runImport(context.Background(), Config{}, nil); code == 0 {
		t.Errorf("got exit code %d for a missing path argument, want non-zero", code)
	}
}

// A path typo is the most likely way this command is used wrongly, so the
// error has to name the file rather than just saying "no such file".
func TestImportFileReportsMissingPath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such-line.csv")

	err := importFile(context.Background(), Wires{}, missing)
	if err == nil {
		t.Fatal("got nil error for a missing file, want an error")
	}
	if !strings.Contains(err.Error(), missing) {
		t.Errorf("error %q does not mention the path %q", err, missing)
	}
}

func TestParseReadingRowRejectsBadDecay(t *testing.T) {
	row := make([]string, 30)
	for i := range row {
		row[i] = "1"
	}
	row[0], row[1] = "line-a-r00001", "line-a"
	row[2], row[6], row[10], row[14] = "e1", "e2", "e3", "e4"
	row[18], row[19] = "2026-08-17T00:00:00Z", "dipole-dipole"
	row[28] = "0.1;0.2;0.3" // only 3 gates

	_, reason := parseReadingRow(row)
	if reason == "" {
		t.Fatal("expected rejection for short decay_curve")
	}
	if !strings.Contains(reason, "10 gates") {
		t.Errorf("reason %q should mention 10 gates", reason)
	}
}
