package database

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	Id        int       `json:"id"`
	Uuid      uuid.UUID `json:"uuid"` // all objects/entities in system have globally unique id
	Type      string    `json:"type"` // user_create, user_login, user_logout, user_delete
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}
