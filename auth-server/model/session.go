package model

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        string    `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}
