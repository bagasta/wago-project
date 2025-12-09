package model

import "time"

type Analytics struct {
	ID                  int64     `json:"id"`
	SessionID           string    `json:"session_id"`
	MessageID           string    `json:"message_id"`
	FromNumber          string    `json:"from_number"`
	MessageType         string    `json:"message_type"`
	IsGroup             bool      `json:"is_group"`
	IsMention           bool      `json:"is_mention"`
	WebhookSent         bool      `json:"webhook_sent"`
	WebhookSuccess      bool      `json:"webhook_success"`
	WebhookResponseTime int       `json:"webhook_response_time_ms"`
	WebhookStatusCode   int       `json:"webhook_status_code"`
	ErrorMessage        string    `json:"error_message"`
	CreatedAt           time.Time `json:"created_at"`
}

type MessageLog struct {
	ID              int64     `json:"id"`
	SessionID       string    `json:"session_id"`
	Direction       string    `json:"direction"` // incoming, outgoing
	FromNumber      string    `json:"from_number"`
	ToNumber        string    `json:"to_number"`
	MessageType     string    `json:"message_type"`
	Content         string    `json:"content"`
	MediaURL        string    `json:"media_url"`
	GroupID         string    `json:"group_id"`
	GroupName       string    `json:"group_name"`
	IsGroup         bool      `json:"is_group"`
	QuotedMessageID string    `json:"quoted_message_id"`
	Timestamp       time.Time `json:"timestamp"`
}

type SessionAnalytics struct {
	TotalMessages      int         `json:"total_messages"`
	IncomingMessages   int         `json:"incoming_messages"`
	OutgoingMessages   int         `json:"outgoing_messages"`
	WebhookSuccessRate float64     `json:"webhook_success_rate"`
	AvgResponseTime    float64     `json:"avg_response_time"`
	LastActive         *time.Time  `json:"last_active"`
	GroupMentions      int         `json:"group_mentions"`
	DailyStats         []DailyStat `json:"daily_stats"`
}

type DailyStat struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
