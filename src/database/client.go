package database

import (
	"database/sql"
	"log"
	"os"
)

type SqlClient interface{}

type dbClient struct {
	connPool *sql.DB // safe for multiple goroutines
}

func NewDBClient() *dbClient {
	return &dbClient{}
}

func (c *dbClient) Connect() {
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
func (c *dbClient) Close() {
	c.connPool.Close()
}
