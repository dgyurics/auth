package repository

import (
	"database/sql"
	"log"
	"os"
)

// dbClient is a wrapper around sql.DB
type DbClient struct {
	connPool *sql.DB // safe for multiple goroutines
}

func NewDBClient() *DbClient {
	return &DbClient{}
}

func (c *DbClient) Connect() {
	connStr := os.Getenv("PG_CONN_STR")
	if connStr == "" {
		log.Fatal("PG_CONN_STR not set")
		return
	}
	connection, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
		return
	}
	c.connPool = connection
}

// The returned DB is safe for concurrent use by multiple goroutines and maintains
// its own pool of idle connections. Thus, the Open function should be called just once.
// It is rarely necessary to close a DB.
func (c *DbClient) Close() {
	c.connPool.Close()
}
