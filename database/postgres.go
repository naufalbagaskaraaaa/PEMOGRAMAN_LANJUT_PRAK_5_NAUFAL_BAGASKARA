package database

import (
	"context"
	"fmt"
	"time"

	"latihan-fiber/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	host, err := config.RequiredEnv("DB_HOST")
	if err != nil {
		return nil, err
	}
	port, err := config.RequiredEnv("DB_PORT")
	if err != nil {
		return nil, err
	}
	user, err := config.RequiredEnv("DB_USER")
	if err != nil {
		return nil, err
	}
	password, err := config.RequiredEnv("DB_PASSWORD")
	if err != nil {
		return nil, err
	}
	dbName, err := config.RequiredEnv("DB_NAME")
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("membuat pool postgres: %w", err)
	}

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctxPing); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database postgres: %w", err)
	}

	return pool, nil
}
