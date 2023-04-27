package model

import (
	"time"

	"github.com/google/uuid"
)

// EventType is a custom type created to enforce a specific set of values
// that can be used to represent different types of events in the system.
type EventType string

// Values for EventType
const (
	LoggedIn       EventType = "logged_in"
	LoggedOut      EventType = "logged_out"
	AccountCreated EventType = "account_created"
)

// Event represents an immutable event that has occurred in the system.
type Event struct {
	ID        int64     `json:"id"`
	UUID      uuid.UUID `json:"uuid"`
	Type      EventType `json:"type"`
	Body      []byte    `json:"body"` // TODO change to map[string]interface{}?
	CreatedAt time.Time `json:"created_at"`
}
