package model

import "github.com/google/uuid"

// User represents a user account.
type User struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Password string    `json:"password"`
}

// OmitPassword creates a copy of the user with the password field set to ""
func OmitPassword(user *User) *User {
	return &User{
		ID:       user.ID,
		Username: user.Username,
		Password: "",
	}
}
