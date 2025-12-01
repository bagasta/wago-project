package model

import (
	"time"
)

type User struct {
	ID        string     `json:"id"`
	PIN       string     `json:"pin"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}
