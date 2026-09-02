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
