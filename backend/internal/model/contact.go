package model

import "time"

type Contact struct {
	PhoneNumber  string    `json:"phone_number"`
	LastActive   time.Time `json:"last_active"`
	MessageCount int       `json:"message_count"`
}
