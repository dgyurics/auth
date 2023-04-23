package repository

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/dgyurics/auth/src/config"
	_ "github.com/lib/pq" // driver for PostgreSQL that provides an implementation of the database/sql package
)

// DbClient is a wrapper around the sql.DB struct
// It is used to connect to the database and execute queries
// Safe for concurrent use by multiple goroutines
type DbClient struct {
	connPool *sql.DB
}

// NewDBClient returns a new instance of DbClient
func NewDBClient() *DbClient {
	return &DbClient{}
}

// Connect establishes a connection to PostgreSQL database
// when provided with a valid config
// FIXME unable to specify schema name via search_path=auth
// https://github.com/go-pg/pg/issues/351#issuecomment-474875596
func (c *DbClient) Connect(config config.PostgreSQL) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s search_path=%s",
		config.Host, config.Port, config.User, config.Password, config.Dbname, config.Sslmode, config.FallbackApplication)
	connection, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = connection.Ping()
	if err != nil {
		log.Fatal(err)
		return
	}
	c.connPool = connection
}

// Close closes the database and prevents new queries from starting.
// Close then waits for all queries that have started processing on the server
// to finish.
//
// It is rare to Close a DB, as the DB handle is meant to be
// long-lived and shared between many goroutines.
func (c *DbClient) Close() {
	if err := c.connPool.Close(); err != nil {
		log.Fatal(err)
	}
}
