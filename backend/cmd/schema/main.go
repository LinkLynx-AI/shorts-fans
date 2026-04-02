package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/LinkLynx-AI/shorts-fans/backend/internal/config"
	"github.com/LinkLynx-AI/shorts-fans/backend/internal/dbschema"
)

const (
	defaultMigrationDir = "db/migrations"
	defaultOutputPath   = "db/schema.generated.yaml"
	commandTimeout      = 2 * time.Minute
)

func main() {
	cfg := config.Load()

	options, err := parseOptions(os.Args[1:], cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	if err := dbschema.Generate(ctx, options); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseOptions(args []string, cfg config.Config) (dbschema.Options, error) {
	fs := flag.NewFlagSet("schema", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	migrationDir := fs.String("migrations", defaultMigrationDir, "path to migration directory")
	outputPath := fs.String("out", defaultOutputPath, "path to generated schema yaml")

	if err := fs.Parse(args); err != nil {
		return dbschema.Options{}, fmt.Errorf("usage: go run ./cmd/schema [-migrations dir] [-out path]")
	}
	if fs.NArg() != 0 {
		return dbschema.Options{}, fmt.Errorf("usage: go run ./cmd/schema [-migrations dir] [-out path]")
	}

	return dbschema.Options{
		AppEnv:       cfg.AppEnv,
		PostgresDSN:  cfg.PostgresDSN,
		MigrationDir: *migrationDir,
		OutputPath:   *outputPath,
	}, nil
}
