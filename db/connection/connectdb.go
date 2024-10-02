package connection

import (
	"bookmark/util"
	"context"
	"time"

	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB() *pgxpool.Pool {
	config, err := util.LoadConfig(".")
	if err != nil {
		panic(err)
	}
	poolConfig, err := pgxpool.ParseConfig(config.DBString)
	if err != nil {
		log.Fatalf("Unable to parse database connection string: %v", err)
	}
	// pgxLogger := pgxzerolog.NewLogger(logger)
	// poolConfig.ConnConfig.Tracer = pgxLogger
	// poolConfig.ConnConfig.Config = pgxLogger
	poolConfig.MaxConns = 150
	poolConfig.MaxConnIdleTime = 1 * time.Second
	poolConfig.MaxConnLifetime = 30 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}

	err = dbPool.Ping(ctx)
	if err != nil {
		log.Fatalf("Unable to ping the database: %v", err)
	}

	return dbPool
}
