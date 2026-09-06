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
	// Ten connections per pool. Both the API and the BFF open one pool per task
	// against the same database, so budget max_connections for
	// (API tasks + BFF tasks) * 10 plus headroom for migrations before scaling.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	// Retire idle sessions quickly and periodically replace long-lived sessions.
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}
