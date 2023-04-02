package model

type User struct {
	Id       string `json:"id"` // FIXME this should be google/uuid.UUID
	Username string `json:"username"`
	Password string `json:"password"`
}
