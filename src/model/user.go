package model

import "github.com/google/uuid"

// User represents a user account.
type User struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Password string    `json:"password"`
}
