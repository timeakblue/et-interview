package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

const importUsage = `usage: go run . import <csv-path>...

Loads survey readings from CSV into the database named by DB_CONN.

	go run . import ../data/line-a.csv
	go run . import ../data/line-*.csv
`

// runImport is the `import` subcommand. The database is opened and migrated
// for you; parsing, validating, and persisting the rows is the exercise.
func runImport(ctx context.Context, conf Config, paths []string) int {
	if len(paths) == 0 {
		fmt.Fprint(os.Stderr, importUsage)
		return 2
	}

	wires, err := WireUp(ctx, conf)
	if err != nil {
		slog.Error("failed to setup app", slog.String("error", err.Error()))
		return 1
	}
	defer func() {
		if err := wires.DB.Close(); err != nil {
			slog.Warn("failed to cleanly close db", slog.String("error", err.Error()))
		}
	}()

	for _, path := range paths {
		if err := importFile(ctx, wires, path); err != nil {
			slog.Error("import failed", slog.String("path", path), slog.String("error", err.Error()))
			return 1
		}
	}

	return 0
}

// importFile loads one CSV into the database.
//
// TODO: this is yours to build. The file is opened for you because the error
// message for a bad path is worth getting right and is not what we are asking
// you to write. What happens next — how rows are parsed, which ones are
// rejected, and whether a bad row fails the file or is reported and skipped —
// is a decision we want to see you make.
func importFile(ctx context.Context, w Wires, path string) error {
	f, err := os.Open(path)
	if err != nil {
		// os.Open's error already names the path.
		return fmt.Errorf("read csv: %w", err)
	}
	defer f.Close()

	return fmt.Errorf("import %s: not implemented", path)
}
