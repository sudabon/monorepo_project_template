package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

// Open configures the pool without requiring a live database. This allows the
// process and shallow health check to stay up during a database outage.
func Open(url string) (*sql.DB, error) {
	config, err := pgx.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("invalid database configuration")
	}
	config.ConnectTimeout = 3 * time.Second
	db := stdlib.OpenDB(*config)
	// Ten connections per task leaves capacity for other tasks and migrations;
	// size the DB max_connections for the maximum ECS task count before scaling.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	// Retire idle sessions quickly and periodically replace long-lived sessions.
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}
