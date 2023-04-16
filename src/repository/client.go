package repository

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/dgyurics/auth/src/config"
	_ "github.com/lib/pq"
)

type DbClient struct {
	connPool *sql.DB // safe for multiple goroutines
}

func NewDBClient() *DbClient {
	return &DbClient{}
}

// FIXME unable to specify schema name via search_path=auth
// https://github.com/go-pg/pg/issues/351#issuecomment-474875596
func (c *DbClient) Connect(config config.PostgreSql) {
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

// The returned DB is safe for concurrent use by multiple goroutines and maintains
// its own pool of idle connections. Thus, the Open function should be called just once.
// It is rarely necessary to close a DB.
func (c *DbClient) Close() {
	c.connPool.Close()
}
