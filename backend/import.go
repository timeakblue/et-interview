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
// Bad data rows are skipped and reported; the rest of the file still imports.
func importFile(ctx context.Context, w Wires, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("read csv: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.ReuseRecord = true
	r.FieldsPerRecord = -1 // validate column count ourselves so we can skip bad rows

	header, err := r.Read()
	if err != nil {
		return fmt.Errorf("read csv header: %w", err)
	}
	if len(header) < 30 {
		return fmt.Errorf("expected 30 columns, got %d", len(header))
	}

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

	accepted, skipped := 0, 0
	lineNo := 1 // header is line 1

	for {
		lineNo++
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			skipped++
			slog.Warn("skipping csv row",
				slog.String("path", path),
				slog.Int("line", lineNo),
				slog.String("reason", err.Error()),
			)
			continue
		}

		parsed, reason := parseReadingRow(row)
		if reason != "" {
			skipped++
			id := ""
			if len(row) > 0 {
				id = strings.TrimSpace(row[0])
			}
			slog.Warn("skipping invalid row",
				slog.String("path", path),
				slog.Int("line", lineNo),
				slog.String("id", id),
				slog.String("reason", reason),
			)
			continue
		}

		_, err = stmt.ExecContext(
			ctx,
			parsed.id, parsed.lineID,
			parsed.tx1ID, parsed.tx1Lat, parsed.tx1Lon, parsed.tx1Alt,
			parsed.tx2ID, parsed.tx2Lat, parsed.tx2Lon, parsed.tx2Alt,
			parsed.rx1ID, parsed.rx1Lat, parsed.rx1Lon, parsed.rx1Alt,
			parsed.rx2ID, parsed.rx2Lat, parsed.rx2Lon, parsed.rx2Alt,
			parsed.timestamp, parsed.arrayType, parsed.symmetry, parsed.stacks,
			parsed.inputCurrent, parsed.contactResistance,
			parsed.apparentResistivity, parsed.apparentResistivityErr,
			parsed.chargeability, parsed.chargeabilityErr,
			parsed.decayCurve, parsed.qcReview,
		)
		if err != nil {
			skipped++
			slog.Warn("skipping row (db error)",
				slog.String("path", path),
				slog.Int("line", lineNo),
				slog.String("id", parsed.id),
				slog.String("reason", err.Error()),
			)
			continue
		}
		accepted++
	} 

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	slog.Info("imported survey line",
		slog.String("path", path),
		slog.Int("accepted", accepted),
		slog.Int("skipped", skipped),
	)
	return nil
}

type readingRow struct {
	id, lineID                                           string
	tx1ID                                                string
	tx1Lat, tx1Lon, tx1Alt                               float64
	tx2ID                                                string
	tx2Lat, tx2Lon, tx2Alt                               float64
	rx1ID                                                string
	rx1Lat, rx1Lon, rx1Alt                               float64
	rx2ID                                                string
	rx2Lat, rx2Lon, rx2Alt                               float64
	timestamp, arrayType                                 string
	symmetry, stacks                                     int
	inputCurrent, contactResistance                      float64
	apparentResistivity, apparentResistivityErr          float64
	chargeability, chargeabilityErr                      float64
	decayCurve, qcReview                                 string
}

// parseReadingRow validates and parses one CSV data row.
// On failure it returns a non-empty reason string.
func parseReadingRow(row []string) (readingRow, string) {
	var out readingRow
	if len(row) < 30 {
		return out, fmt.Sprintf("expected 30 columns, got %d", len(row))
	}

	require := func(idx int, name string) (string, string) {
		v := strings.TrimSpace(row[idx])
		if v == "" {
			return "", fmt.Sprintf("missing %s", name)
		}
		return v, ""
	}

	parseFloat := func(idx int, name string) (float64, string) {
		raw := strings.TrimSpace(row[idx])
		if raw == "" {
			return 0, fmt.Sprintf("missing %s", name)
		}
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return 0, fmt.Sprintf("invalid %s %q", name, raw)
		}
		return v, ""
	}

	parseInt := func(idx int, name string) (int, string) {
		raw := strings.TrimSpace(row[idx])
		if raw == "" {
			return 0, fmt.Sprintf("missing %s", name)
		}
		v, err := strconv.Atoi(raw)
		if err != nil {
			return 0, fmt.Sprintf("invalid %s %q", name, raw)
		}
		return v, ""
	}

	var reason string
	if out.id, reason = require(0, "id"); reason != "" {
		return out, reason
	}
	if out.lineID, reason = require(1, "line_id"); reason != "" {
		return out, reason
	}

	if out.tx1ID, reason = require(2, "tx1_id"); reason != "" {
		return out, reason
	}
	if out.tx1Lat, reason = parseFloat(3, "tx1_lat"); reason != "" {
		return out, reason
	}
	if out.tx1Lon, reason = parseFloat(4, "tx1_lon"); reason != "" {
		return out, reason
	}
	if out.tx1Alt, reason = parseFloat(5, "tx1_alt"); reason != "" {
		return out, reason
	}

	if out.tx2ID, reason = require(6, "tx2_id"); reason != "" {
		return out, reason
	}
	if out.tx2Lat, reason = parseFloat(7, "tx2_lat"); reason != "" {
		return out, reason
	}
	if out.tx2Lon, reason = parseFloat(8, "tx2_lon"); reason != "" {
		return out, reason
	}
	if out.tx2Alt, reason = parseFloat(9, "tx2_alt"); reason != "" {
		return out, reason
	}

	if out.rx1ID, reason = require(10, "rx1_id"); reason != "" {
		return out, reason
	}
	if out.rx1Lat, reason = parseFloat(11, "rx1_lat"); reason != "" {
		return out, reason
	}
	if out.rx1Lon, reason = parseFloat(12, "rx1_lon"); reason != "" {
		return out, reason
	}
	if out.rx1Alt, reason = parseFloat(13, "rx1_alt"); reason != "" {
		return out, reason
	}

	if out.rx2ID, reason = require(14, "rx2_id"); reason != "" {
		return out, reason
	}
	if out.rx2Lat, reason = parseFloat(15, "rx2_lat"); reason != "" {
		return out, reason
	}
	if out.rx2Lon, reason = parseFloat(16, "rx2_lon"); reason != "" {
		return out, reason
	}
	if out.rx2Alt, reason = parseFloat(17, "rx2_alt"); reason != "" {
		return out, reason
	}

	if out.timestamp, reason = require(18, "timestamp"); reason != "" {
		return out, reason
	}
	if out.arrayType, reason = require(19, "array_type"); reason != "" {
		return out, reason
	}
	if out.symmetry, reason = parseInt(20, "symmetry"); reason != "" {
		return out, reason
	}
	if out.stacks, reason = parseInt(21, "stacks"); reason != "" {
		return out, reason
	}
	if out.inputCurrent, reason = parseFloat(22, "input_current"); reason != "" {
		return out, reason
	}
	if out.contactResistance, reason = parseFloat(23, "contact_resistance"); reason != "" {
		return out, reason
	}
	if out.apparentResistivity, reason = parseFloat(24, "apparent_resistivity"); reason != "" {
		return out, reason
	}
	if out.apparentResistivityErr, reason = parseFloat(25, "apparent_resistivity_err"); reason != "" {
		return out, reason
	}
	if out.chargeability, reason = parseFloat(26, "chargeability"); reason != "" {
		return out, reason
	}
	if out.chargeabilityErr, reason = parseFloat(27, "chargeability_err"); reason != "" {
		return out, reason
	}

	if out.decayCurve, reason = require(28, "decay_curve"); reason != "" {
		return out, reason
	}
	parts := strings.Split(out.decayCurve, ";")
	if len(parts) != 10 {
		return out, fmt.Sprintf("decay_curve must have 10 gates, got %d", len(parts))
	}
	for i, p := range parts {
		if _, err := strconv.ParseFloat(strings.TrimSpace(p), 64); err != nil {
			return out, fmt.Sprintf("invalid decay_curve gate %d %q", i+1, p)
		}
	}

	out.qcReview = strings.TrimSpace(row[29]) // may be empty
	return out, ""
}
