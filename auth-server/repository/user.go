package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/dgyurics/auth/auth-server/model"
)

// UserRepository is an interface for interacting with the user table
type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	ExistsUser(ctx context.Context, username string) bool
	GetUser(ctx context.Context, user *model.User) error
}

type userRepository struct {
	*DbClient
	stmtInsertEvent          *sql.Stmt // Prepared statement for inserting into auth.event
	stmtInsertUser           *sql.Stmt // Prepared statement for inserting into auth.user
	stmtSelectUserByUsername *sql.Stmt // Prepared statement for selecting a user by username
	stmtSelectUserByID       *sql.Stmt // Prepared statement for selecting a user by ID
}

// NewUserRepository creates a new user repository
func NewUserRepository(c *DbClient) UserRepository {
	repo := &userRepository{
		DbClient: c,
	}
	repo.prepareStatements()
	return repo
}

func (r *userRepository) ExistsUser(ctx context.Context, username string) bool {
	if err := r.GetUser(ctx, &model.User{Username: username}); err != nil {
		return false
	}
	return true
}

func (r *userRepository) GetUser(ctx context.Context, user *model.User) error {
	var row *sql.Row
	if user.Username != "" {
		row = r.stmtSelectUserByUsername.QueryRowContext(ctx, user.Username)
	} else {
		row = r.stmtSelectUserByID.QueryRowContext(ctx, user.ID.String())
	}
	return row.Scan(&user.ID, &user.Username, &user.Password)
}

func (r *userRepository) CreateUser(ctx context.Context, user *model.User) error {
	tx, err := r.connPool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback() // FIXME error handling
			return
		}
		err = tx.Commit()
	}()

	// stringify user for event body
	userEncoded, err := json.Marshal(model.OmitPassword(user))
	if err != nil {
		return err
	}

	_, err = r.stmtInsertEvent.Exec(user.ID, model.AccountCreated, userEncoded)
	if err != nil {
		return err
	}

	if _, err = r.stmtInsertUser.Exec(user.ID, user.Username, user.Password); err != nil {
		return err
	}

	return nil
}

// https://go.dev/doc/database/prepared-statements
func (r *userRepository) Close(ctx context.Context) {
	if err := r.stmtInsertEvent.Close(); err != nil {
		log.Println(err)
	}
	if err := r.stmtInsertUser.Close(); err != nil {
		log.Println(err)
	}
	if err := r.stmtSelectUserByUsername.Close(); err != nil {
		log.Println(err)
	}
	if err := r.stmtSelectUserByID.Close(); err != nil {
		log.Println(err)
	}
}

// Prepare the necessary SQL statements
func (r *userRepository) prepareStatements() {
	var err error
	r.stmtInsertEvent, err = r.connPool.Prepare("INSERT INTO auth.event (uuid, type, body) VALUES ($1, $2, $3)")
	if err != nil {
		log.Fatal(err)
	}

	r.stmtInsertUser, err = r.connPool.Prepare("INSERT INTO auth.user (id, username, password) VALUES ($1, $2, $3)")
	if err != nil {
		log.Fatal(err)
	}

	r.stmtSelectUserByUsername, err = r.connPool.Prepare("SELECT id, username, password FROM auth.user WHERE username = $1")
	if err != nil {
		log.Fatal(err)
	}

	r.stmtSelectUserByID, err = r.connPool.Prepare("SELECT id, username, password FROM auth.user WHERE id = $1")
	if err != nil {
		log.Fatal(err)
	}
}
