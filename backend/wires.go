package main

import (
	"context"
	"fmt"
	"log/slog"

	"et-interview/internal/db"
	"et-interview/internal/demo"
)

type Wires struct {
	DB      db.DB
	Demo    demo.Store
	WebRoot string
}

func WireUp(ctx context.Context, conf Config) (w Wires, err error) {
	w.DB, err = db.Connect(ctx, conf.DBConn)
	if err != nil {
		return
	}

	// Applied on every startup; see internal/db/schema.sql.
	if err = db.Migrate(ctx, w.DB); err != nil {
		_ = w.DB.Close()
		err = fmt.Errorf("migrate: %w", err)
		return
	}

	w.Demo = demo.NewStore(w.DB)
	w.WebRoot = conf.WebRoot
	return
}

// Run cleanup functions when app shuts down
func (w Wires) Run(ctx context.Context) error {
	<-ctx.Done()
	slog.Info("wires cleanup")

	if dbCloseErr := w.DB.Close(); dbCloseErr != nil {
		slog.Warn("failed to cleanly close db", slog.String("error", dbCloseErr.Error()))
	}

	return nil
}
