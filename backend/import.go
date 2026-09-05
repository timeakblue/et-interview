package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
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
// importFile loads one CSV into the database.
func importFile(ctx context.Context, w Wires, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("read csv: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)

	// Read header row
	header, err := r.Read()
	if err != nil {
		return fmt.Errorf("read csv header: %w", err)
	}
	if len(header) < 30 {
		return fmt.Errorf("expected 30 columns, got %d", len(header))
	}

	// Begin transaction directly on the Write pool
	tx, err := w.DB.Write.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO readings (
			id, line_id,
			tx1_id, tx1_lat, tx1_lon, tx1_alt,
			tx2_id, tx2_lat, tx2_lon, tx2_alt,
			rx1_id, rx1_lat, rx1_lon, rx1_alt,
			rx2_id, rx2_lat, rx2_lon, rx2_alt,
			timestamp, array_type, symmetry, stacks,
			input_current, contact_resistance,
			apparent_resistivity, apparent_resistivity_err,
			chargeability, chargeability_err,
			decay_curve, qc_review
		) VALUES (
			?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?,
			?, ?,
			?, ?,
			?, ?
		)
		ON CONFLICT(id) DO UPDATE SET
			line_id = excluded.line_id,
			tx1_id = excluded.tx1_id, tx1_lat = excluded.tx1_lat, tx1_lon = excluded.tx1_lon, tx1_alt = excluded.tx1_alt,
			tx2_id = excluded.tx2_id, tx2_lat = excluded.tx2_lat, tx2_lon = excluded.tx2_lon, tx2_alt = excluded.tx2_alt,
			rx1_id = excluded.rx1_id, rx1_lat = excluded.rx1_lat, rx1_lon = excluded.rx1_lon, rx1_alt = excluded.rx1_alt,
			rx2_id = excluded.rx2_id, rx2_lat = excluded.rx2_lat, rx2_lon = excluded.rx2_lon, rx2_alt = excluded.rx2_alt,
			timestamp = excluded.timestamp, array_type = excluded.array_type, symmetry = excluded.symmetry, stacks = excluded.stacks,
			input_current = excluded.input_current, contact_resistance = excluded.contact_resistance,
			apparent_resistivity = excluded.apparent_resistivity, apparent_resistivity_err = excluded.apparent_resistivity_err,
			chargeability = excluded.chargeability, chargeability_err = excluded.chargeability_err,
			decay_curve = excluded.decay_curve, qc_review = excluded.qc_review
	`)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	count := 0
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read csv row %d: %w", count+1, err)
		}

		parseFloat := func(idx int) float64 {
			val, _ := strconv.ParseFloat(strings.TrimSpace(row[idx]), 64)
			return val
		}

		parseInt := func(idx int) int {
			val, _ := strconv.Atoi(strings.TrimSpace(row[idx]))
			return val
		}

		id := strings.TrimSpace(row[0])
		lineID := strings.TrimSpace(row[1])

		tx1ID, tx1Lat, tx1Lon, tx1Alt := strings.TrimSpace(row[2]), parseFloat(3), parseFloat(4), parseFloat(5)
        tx2ID, tx2Lat, tx2Lon, tx2Alt := strings.TrimSpace(row[6]), parseFloat(7), parseFloat(8), parseFloat(9)
        rx1ID, rx1Lat, rx1Lon, rx1Alt := strings.TrimSpace(row[10]), parseFloat(11), parseFloat(12), parseFloat(13)
        rx2ID, rx2Lat, rx2Lon, rx2Alt := strings.TrimSpace(row[14]), parseFloat(15), parseFloat(16), parseFloat(17)

		timestamp := strings.TrimSpace(row[18])
		arrayType := strings.TrimSpace(row[19])
		symmetry := parseInt(20)
		stacks := parseInt(21)
		inputCurrent := parseFloat(22)
		contactResistance := parseFloat(23)

		apparentResistivity := parseFloat(24)
		apparentResistivityErr := parseFloat(25)
		chargeability := parseFloat(26)
		chargeabilityErr := parseFloat(27)

		decayCurve := strings.TrimSpace(row[28])
		qcReview := strings.TrimSpace(row[29])

		_, err = stmt.ExecContext(
			ctx,
			id, lineID,
			tx1ID, tx1Lat, tx1Lon, tx1Alt,
			tx2ID, tx2Lat, tx2Lon, tx2Alt,
			rx1ID, rx1Lat, rx1Lon, rx1Alt,
			rx2ID, rx2Lat, rx2Lon, rx2Alt,
			timestamp, arrayType, symmetry, stacks,
			inputCurrent, contactResistance,
			apparentResistivity, apparentResistivityErr,
			chargeability, chargeabilityErr,
			decayCurve, qcReview,
		)
		if err != nil {
			return fmt.Errorf("exec row %s: %w", id, err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	slog.Info("imported survey line successfully", slog.String("path", path), slog.Int("records", count))
	return nil
}