package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"

	"github.com/kelseyhightower/envconfig"
	"github.com/quells/grace"
	"github.com/quells/grace/pkg/httpserver"
)

type Config struct {
	HTTPAddr string `envconfig:"HTTP_ADDR" default:""`
	HTTPPort int    `envconfig:"HTTP_PORT" default:"8080"`
	WebRoot  string `envconfig:"WEB_ROOT" default:"/app/web"`
	DBConn   string `envconfig:"DB_CONN" default:"qc.db"`
}

const usage = `usage: go run . [command]

Commands:
	(none)    start the HTTP server on HTTP_PORT (default 8080)
	import    load survey readings from CSV into the database

Run "go run . import" for the import command's own usage.
`

func main() {
	ctx := context.Background()
	var conf Config
	if err := envconfig.Process("", &conf); err != nil {
		slog.Error("failed to parse environment variables", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// No subcommand means "serve", so `go run .` and the container's default
	// command keep working.
	if len(os.Args) > 1 {
		switch cmd := os.Args[1]; cmd {
		case "import":
			os.Exit(runImport(ctx, conf, os.Args[2:]))
		default:
			fmt.Fprintf(os.Stderr, "unknown command %q\n\n%s", cmd, usage)
			os.Exit(2)
		}
	}

	os.Exit(run(ctx, conf))
}

func run(ctx context.Context, conf Config) int {
	var stop func()
	ctx, stop = grace.StopOnInterrupt(ctx)
	defer stop()

	wires, err := WireUp(ctx, conf)
	if err != nil {
		slog.Error("failed to setup app", slog.String("error", err.Error()))
		return 1
	}

	httpAddr := net.JoinHostPort(conf.HTTPAddr, strconv.Itoa(conf.HTTPPort))
	httpHandler := Routes(wires)
	httpServer := httpserver.New(
		httpAddr,
		httpHandler,
		httpserver.WithLogger(slog.Default()),
	)

	err = grace.Run(ctx, httpServer, wires)
	if err != nil {
		slog.Error("server error", slog.String("error", err.Error()))
		return 1
	}

	slog.Info("goodbye")
	return 0
}
