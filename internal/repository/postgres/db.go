package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mmd-moradi/goup/configs"
	"github.com/rs/zerolog"
)

func NewDBPool(cfg *configs.DatabaseConfig, logger zerolog.Logger) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Passsword,
		cfg.Host,
		cfg.Port,
		cfg.DbName,
		cfg.SSLMode,
	)
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}
	poolConfig.MaxConns = 25

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
