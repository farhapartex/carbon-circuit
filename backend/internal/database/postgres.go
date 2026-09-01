package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const transactionPoolingForbidsPreparedStatements = true

type Options struct {
	DSN             string
	Schema          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	AcquireTimeout  time.Duration
}

func Open(ctx context.Context, options Options) (*gorm.DB, error) {
	dialector := postgres.New(postgres.Config{
		DSN:                  options.DSN,
		PreferSimpleProtocol: transactionPoolingForbidsPreparedStatements,
	})

	database, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   options.Schema + ".",
			SingularTable: false,
		},
		SkipDefaultTransaction: true,
		PrepareStmt:            false,
		TranslateError:         true,
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	pool, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("resolve connection pool: %w", err)
	}

	pool.SetMaxOpenConns(options.MaxOpenConns)
	pool.SetMaxIdleConns(options.MaxIdleConns)
	pool.SetConnMaxLifetime(options.ConnMaxLifetime)
	pool.SetConnMaxIdleTime(options.ConnMaxIdleTime)

	pingCtx, cancel := context.WithTimeout(ctx, options.AcquireTimeout)
	defer cancel()

	if err := pool.PingContext(pingCtx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return database, nil
}

func Close(database *gorm.DB) error {
	pool, err := database.DB()
	if err != nil {
		return err
	}
	return pool.Close()
}
