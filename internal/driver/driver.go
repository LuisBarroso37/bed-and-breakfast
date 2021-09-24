package driver

import (
	"database/sql"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// Holds the database connection pool
type DB struct {
	SQL *sql.DB
}

// Connection options
const maxConnections = 10
const maxIdleConnections = 5
const maxConnectionLifetime = 5 * time.Minute

var dbConn = &DB{}

// Create `DB` struct holding connection pool with set options 
func ConnectSQL(connectionString string) (*DB, error) {
	// Create connection pool
	pool, err := CreateConnectionPool(connectionString)
	if err != nil {
		panic(err)
	}

	// Set connection pool options
	pool.SetMaxOpenConns(maxConnections)
	pool.SetMaxIdleConns(maxIdleConnections)
	pool.SetConnMaxLifetime(maxConnectionLifetime)

	// Store connection pool in `DB struct`
	dbConn.SQL = pool

	// Test database connection pool
	err = testConnectionPool(pool)
	if err != nil {
		return nil, err
	}

	return dbConn, nil
}

// Test database connection pool by pinging it
func testConnectionPool(pool *sql.DB) error {
	err := pool.Ping()
	if err != nil {
		return err
	}

	return nil
}

// Create a connection pool
func CreateConnectionPool(connectionString string) (*sql.DB, error) {
	// Create database connection pool
	pool, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, err
	}

	// Test database connection pool
	err = pool.Ping()
	if err != nil {
		return nil, err
	}

	return pool, nil
}