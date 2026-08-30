package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/carboncircuit/backend/internal/migrate"
	"github.com/carboncircuit/backend/services/identity-service/migrations"
)

const usage = `usage: migrate <up|down|version|force> [argument]

  up               apply every pending migration
  down [steps]     roll back all migrations, or the given number of steps
  version          report the current version and whether it is dirty
  force <version>  mark a version as applied without running it, to clear a dirty state
`

func main() {
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(2)
	}

	dsn := os.Getenv("DATABASE_MIGRATION_DSN")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_MIGRATION_DSN is required and must point at Postgres directly, never PgBouncer")
		os.Exit(2)
	}

	schema := os.Getenv("DATABASE_SCHEMA")
	if schema == "" {
		schema = "identity"
	}

	options := migrate.Options{DSN: dsn, Schema: schema, Migrations: migrations.Files}

	status, err := dispatch(flag.Arg(0), flag.Arg(1), options)
	if err != nil {
		fmt.Fprintln(os.Stderr, "migration failed:", err)
		os.Exit(1)
	}

	report(flag.Arg(0), schema, status)
}

func dispatch(command, argument string, options migrate.Options) (migrate.Status, error) {
	switch command {
	case "up":
		return migrate.Up(options)
	case "down":
		steps := 0
		if argument != "" {
			parsed, err := strconv.Atoi(argument)
			if err != nil {
				return migrate.Status{}, fmt.Errorf("steps must be a number, got %q", argument)
			}
			steps = parsed
		}
		return migrate.Down(options, steps)
	case "version":
		return migrate.Version(options)
	case "force":
		parsed, err := strconv.Atoi(argument)
		if err != nil {
			return migrate.Status{}, fmt.Errorf("force requires a version number, got %q", argument)
		}
		return migrate.Force(options, parsed)
	default:
		return migrate.Status{}, fmt.Errorf("unknown command %q", command)
	}
}

func report(command, schema string, status migrate.Status) {
	state := "clean"
	if status.Dirty {
		state = "DIRTY, resolve with: migrate force <version>"
	}

	changed := "no change"
	if status.Applied {
		changed = "changed"
	}

	fmt.Printf("%s: schema=%s version=%d %s (%s)\n", command, schema, status.Version, state, changed)
}
