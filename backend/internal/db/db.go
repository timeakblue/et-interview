package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime"

	_ "modernc.org/sqlite"
)

// DB holds the two connection pools. Write is pinned to one connection because
// SQLite allows one writer at a time; Read is opened read-only and sized for
// concurrency, which WAL mode permits alongside that writer.
//
// For an in-memory database both fields are the same pool: a ":memory:"
// database is per-connection, so a second pool would be a second, empty
// database rather than another view of the same one.
type DB struct {
	Read  *sql.DB
	Write *sql.DB
}

// Close closes both pools, skipping Read when it aliases Write so an in-memory
// handle is not closed twice.
func (d DB) Close() error {
	var errs []error
	if d.Read != nil && d.Read != d.Write {
		errs = append(errs, d.Read.Close())
	}
	if d.Write != nil {
		errs = append(errs, d.Write.Close())
	}
	return errors.Join(errs...)
}

const busyTimeoutMS = 5000

var writePragmas = []string{
	fmt.Sprintf("busy_timeout(%d)", busyTimeoutMS),
	"foreign_keys(1)",
	"journal_mode(WAL)",
}

// readPragmas deliberately omits journal_mode and foreign_keys. journal_mode is
// a persistent property of the database header, set once by the writer;
// foreign_keys constrains writes only. Neither is meaningful on a read-only
// connection.
var readPragmas = []string{
	fmt.Sprintf("busy_timeout(%d)", busyTimeoutMS),
}

// Connect opens the write and read pools for connStr.
//
// PRAGMAs travel in the DSN rather than being executed after opening, because
// a PRAGMA is per-connection state: a statement run against a multi-connection
// pool would configure only whichever connection database/sql handed out.
func Connect(ctx context.Context, connStr string) (DB, error) {
	writeDSN, err := dsn(connStr, nil, writePragmas)
	if err != nil {
		return DB{}, err
	}

	write, err := sql.Open("sqlite", writeDSN)
	if err != nil {
		return DB{}, fmt.Errorf("open write pool: %w", err)
	}
	// SQLite serializes writers anyway; one connection makes that explicit.
	// The idle connection is kept to avoid reconnect churn on a busy server.
	write.SetMaxOpenConns(1)
	write.SetMaxIdleConns(1)

	if err = write.PingContext(ctx); err != nil {
		_ = write.Close()
		return DB{}, fmt.Errorf("open write pool: %w", err)
	}

	// ":memory:" is per-connection, and the empty DSN opens a private temporary
	// file, which is equally per-connection. Either way a second pool would be a
	// separate, empty database, so both fields must share one.
	if connStr == ":memory:" || connStr == "" {
		return DB{Read: write, Write: write}, nil
	}

	// mode=ro turns a write misrouted to the read pool into a loud failure
	// instead of a silent serialization bottleneck.
	readDSN, err := dsn(connStr, map[string]string{"mode": "ro"}, readPragmas)
	if err != nil {
		_ = write.Close()
		return DB{}, err
	}

	read, err := sql.Open("sqlite", readDSN)
	if err != nil {
		_ = write.Close()
		return DB{}, fmt.Errorf("open read pool: %w", err)
	}
	// Idle capacity matches open capacity: database/sql's default idle cap of 2
	// would close and reopen most of the pool under exactly the bursty read load
	// this pool exists to serve, and each reopen re-runs the DSN's PRAGMAs.
	// GOMAXPROCS is cgroup-aware where NumCPU is not.
	readConns := max(4, runtime.GOMAXPROCS(0))
	read.SetMaxOpenConns(readConns)
	read.SetMaxIdleConns(readConns)

	if err = read.PingContext(ctx); err != nil {
		_ = read.Close()
		_ = write.Close()
		return DB{}, fmt.Errorf("open read pool: %w", err)
	}

	return DB{Read: read, Write: write}, nil
}
