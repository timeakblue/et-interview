// Package demo is a throwaway vertical slice: handler -> store -> SQLite ->
// JSON. It exists so you can see the wiring work end to end before you build
// anything real. Delete it, its table in internal/db/schema.sql, and the
// /api/demo route once you are oriented.
package demo

import (
	"context"
	"fmt"

	"et-interview/internal/db"
)

// Store reads demo values. Queries live here rather than in the HTTP handler
// so the handler stays about requests and responses.
type Store struct {
	db db.DB
}

func NewStore(d db.DB) Store {
	return Store{db: d}
}

// Values returns the seeded numbers in position order.
func (s Store) Values(ctx context.Context) ([]float64, error) {
	rows, err := s.db.Read.QueryContext(ctx,
		`SELECT value FROM demo_values ORDER BY position`)
	if err != nil {
		return nil, fmt.Errorf("query demo values: %w", err)
	}
	defer rows.Close()

	// Non-nil so an empty table marshals as [] rather than null.
	values := []float64{}
	for rows.Next() {
		var v float64
		if err := rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("scan demo value: %w", err)
		}
		values = append(values, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read demo values: %w", err)
	}
	return values, nil
}
