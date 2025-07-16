package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/DmitriyKolesnikM8O/subscription-service/config"
	"github.com/DmitriyKolesnikM8O/subscription-service/pkg/utils"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Client interface {
	Close()
	Acquire(ctx context.Context) (*pgxpool.Conn, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Begin(ctx context.Context) (pgx.Tx, error)
	BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	Ping(ctx context.Context) error
}

func NewClient(ctx context.Context, cfg config.StorageConfig, maxAttempts int) (pool *pgxpool.Pool, err error) {
	pgUrl := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	config, err := pgxpool.ParseConfig(pgUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	if cfg.MaxPoolSize > 0 {
		config.MaxConns = int32(cfg.MaxPoolSize)
	}

	err = utils.ConnectWithTries(func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		pool, err = pgxpool.ConnectConfig(ctx, config)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}

		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			return fmt.Errorf("failed to ping database: %w", err)
		}

		return nil
	}, maxAttempts, 5*time.Second)
	if err != nil {
		return nil, err
	}

	return pool, nil

}
