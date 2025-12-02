package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type SessionStatus string

const (
	SessionStatusQR           SessionStatus = "qr"
	SessionStatusConnected    SessionStatus = "connected"
	SessionStatusDisconnected SessionStatus = "disconnected"
)

type DeviceInfo struct {
	Platform           string `json:"platform,omitempty"`
	DeviceManufacturer string `json:"device_manufacturer,omitempty"`
	DeviceModel        string `json:"device_model,omitempty"`
}

// Make DeviceInfo implement sql.Scanner and driver.Valuer
func (d DeviceInfo) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *DeviceInfo) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &d)
}

type Session struct {
	ID                     string        `json:"session_id"`
	UserID                 string        `json:"-"`
	SessionName            string        `json:"session_name"`
	WebhookURL             string        `json:"webhook_url"`
	Status                 SessionStatus `json:"status"`
	QRCode                 string        `json:"qr_code,omitempty"`
	PhoneNumber            string        `json:"phone_number,omitempty"`
	DeviceInfo             *DeviceInfo   `json:"device_info,omitempty"`
	CreatedAt              time.Time     `json:"created_at"`
	UpdatedAt              time.Time     `json:"updated_at"`
	LastConnected          *time.Time    `json:"last_connected,omitempty"`
	UptimeSeconds          int64         `json:"uptime_seconds,omitempty"`
	IsGroupResponseEnabled bool          `json:"is_group_response_enabled"`
}
