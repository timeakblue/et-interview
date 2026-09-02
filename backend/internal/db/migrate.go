package db

import (
	"context"
	_ "embed"
	"fmt"
)

//go:embed schema.sql
var schema string

// Migrate applies schema.sql to the database.
//
// It runs on every startup rather than tracking versions, so every statement
// in schema.sql must be idempotent. That is enough for an exercise of this
// size; a real deployment would want numbered migrations.
func Migrate(ctx context.Context, d DB) error {
	if _, err := d.Write.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	return nil
}
