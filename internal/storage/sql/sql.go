package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type DB struct {
	pool *pgxpool.Pool
	log  *zap.Logger
}

func NewDB(ctx context.Context, dsn string, log *zap.Logger) (*DB, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create a connection pool: %w", err)
	}

	return &DB{
		pool: pool,
		log:  log,
	}, nil
}

func (db *DB) Close() {
	db.pool.Close()
}
