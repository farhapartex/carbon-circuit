package migrate

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

const trackingTable = "schema_migrations"

type Options struct {
	DSN        string
	Schema     string
	Migrations fs.FS
}

type Status struct {
	Version uint
	Dirty   bool
	Applied bool
}

func runner(options Options) (*migrate.Migrate, error) {
	source, err := iofs.New(options.Migrations, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}

	parsed, err := url.Parse(options.DSN)
	if err != nil {
		return nil, fmt.Errorf("parse migration dsn: %w", err)
	}

	query := parsed.Query()
	query.Set("search_path", options.Schema+",public")
	query.Set("x-migrations-table", trackingTable)
	query.Set("x-migrations-table-quoted", "false")
	parsed.RawQuery = query.Encode()

	instance, err := migrate.NewWithSourceInstance("iofs", source, parsed.String())
	if err != nil {
		return nil, fmt.Errorf("open migration runner: %w", err)
	}

	return instance, nil
}

func Up(options Options) (Status, error) {
	instance, err := runner(options)
	if err != nil {
		return Status{}, err
	}
	defer closeQuietly(instance)

	err = instance.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return Status{}, fmt.Errorf("apply migrations: %w", err)
	}

	return currentStatus(instance, !errors.Is(err, migrate.ErrNoChange))
}

func Down(options Options, steps int) (Status, error) {
	instance, err := runner(options)
	if err != nil {
		return Status{}, err
	}
	defer closeQuietly(instance)

	if steps <= 0 {
		err = instance.Down()
	} else {
		err = instance.Steps(-steps)
	}
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return Status{}, fmt.Errorf("roll back migrations: %w", err)
	}

	return currentStatus(instance, !errors.Is(err, migrate.ErrNoChange))
}

func Version(options Options) (Status, error) {
	instance, err := runner(options)
	if err != nil {
		return Status{}, err
	}
	defer closeQuietly(instance)

	return currentStatus(instance, false)
}

func Force(options Options, version int) (Status, error) {
	instance, err := runner(options)
	if err != nil {
		return Status{}, err
	}
	defer closeQuietly(instance)

	if err := instance.Force(version); err != nil {
		return Status{}, fmt.Errorf("force version: %w", err)
	}

	return currentStatus(instance, true)
}

func currentStatus(instance *migrate.Migrate, applied bool) (Status, error) {
	version, dirty, err := instance.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return Status{Applied: applied}, nil
	}
	if err != nil {
		return Status{}, fmt.Errorf("read migration version: %w", err)
	}

	return Status{Version: version, Dirty: dirty, Applied: applied}, nil
}

func closeQuietly(instance *migrate.Migrate) {
	sourceErr, databaseErr := instance.Close()
	_ = sourceErr
	_ = databaseErr
}
