package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type config struct {
	port int
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  time.Duration
	}
}

type application struct {
	config config
	logger *slog.Logger
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 8080, "API port")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://photoapi:secret@localhost:5432/photoapi?sslmode=disable", "Postgres connection string")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "Max open connections to the database")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "Max idle connections to the database")
	flag.DurationVar(&cfg.db.maxIdleTime, "db-max-idle-time", 15*time.Minute, "Max idle time for connections in the database")

	flag.Parse()

	app := fiber.New()

	// logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// application := &application{
	// 	logger: logger,
	// 	config: cfg,
	// }

	// ctx := context.Background()

	pool, err := newPostgresPool(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	// userRepo := pg.NewPgUserRepository(pool)
	// photoRepo := pg.NewPgPhotoRepository(pool)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "ok"})
	})

	log.Fatal(app.Listen(":8080"))

}

func newPostgresPool(cfg config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	poolConfig.MaxConns = int32(cfg.db.maxOpenConns)
	poolConfig.MinConns = int32(cfg.db.maxIdleConns)
	poolConfig.MaxConnLifetime = cfg.db.maxIdleTime

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
