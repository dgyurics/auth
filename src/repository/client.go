package repository

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// dbClient is a wrapper around sql.DB
type DbClient struct {
	connPool *sql.DB // safe for multiple goroutines
}

func NewDBClient() *DbClient {
	return &DbClient{}
}

// FIXME search_path=auth in database url not working
// set schema
// https://github.com/go-pg/pg/issues/351#issuecomment-474875596
func (c *DbClient) Connect() {
	// FIXME use config
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL not set")
		return
	}
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

// The returned DB is safe for concurrent use by multiple goroutines and maintains
// its own pool of idle connections. Thus, the Open function should be called just once.
// It is rarely necessary to close a DB.
func (c *DbClient) Close() {
	c.connPool.Close()
}
