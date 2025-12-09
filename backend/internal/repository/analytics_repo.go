package repository

import (
	"database/sql"
	"wago-backend/internal/model"
)

type AnalyticsRepository struct {
	DB *sql.DB
}

func NewAnalyticsRepository(db *sql.DB) *AnalyticsRepository {
	return &AnalyticsRepository{DB: db}
}

func (r *AnalyticsRepository) LogMessage(log *model.MessageLog) error {
	query := `
		INSERT INTO messages_log (session_id, direction, from_number, to_number, message_type, content, media_url, group_id, group_name, is_group, quoted_message_id, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.DB.Exec(query, log.SessionID, log.Direction, log.FromNumber, log.ToNumber, log.MessageType, log.Content, log.MediaURL, log.GroupID, log.GroupName, log.IsGroup, log.QuotedMessageID, log.Timestamp)
	return err
}

func (r *AnalyticsRepository) LogAnalytics(a *model.Analytics) error {
	query := `
		INSERT INTO analytics (session_id, message_id, from_number, message_type, is_group, is_mention, webhook_sent, webhook_success, webhook_response_time_ms, webhook_status_code, error_message)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.DB.Exec(query, a.SessionID, a.MessageID, a.FromNumber, a.MessageType, a.IsGroup, a.IsMention, a.WebhookSent, a.WebhookSuccess, a.WebhookResponseTime, a.WebhookStatusCode, a.ErrorMessage)
	return err
}

func (r *AnalyticsRepository) GetSessionAnalytics(sessionID string) (*model.SessionAnalytics, error) {
	stats := &model.SessionAnalytics{
		DailyStats: []model.DailyStat{},
	}

	// Total Messages
	err := r.DB.QueryRow("SELECT COUNT(*) FROM messages_log WHERE session_id = $1", sessionID).Scan(&stats.TotalMessages)
	if err != nil {
		return nil, err
	}

	// Incoming
	err = r.DB.QueryRow("SELECT COUNT(*) FROM messages_log WHERE session_id = $1 AND direction = 'incoming'", sessionID).Scan(&stats.IncomingMessages)
	if err != nil {
		return nil, err
	}

	// Outgoing
	err = r.DB.QueryRow("SELECT COUNT(*) FROM messages_log WHERE session_id = $1 AND direction = 'outgoing'", sessionID).Scan(&stats.OutgoingMessages)
	if err != nil {
		return nil, err
	}

	// Webhook Stats
	var totalWebhooks int
	var successWebhooks int
	var totalTime int64
	err = r.DB.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(CASE WHEN webhook_success THEN 1 ELSE 0 END), 0), COALESCE(SUM(webhook_response_time_ms), 0)
		FROM analytics WHERE session_id = $1 AND webhook_sent = true
	`, sessionID).Scan(&totalWebhooks, &successWebhooks, &totalTime)
	if err != nil {
		return nil, err
	}

	if totalWebhooks > 0 {
		stats.WebhookSuccessRate = float64(successWebhooks) / float64(totalWebhooks) * 100
		stats.AvgResponseTime = float64(totalTime) / float64(totalWebhooks)
	}

	// Group Mentions
	err = r.DB.QueryRow("SELECT COUNT(*) FROM analytics WHERE session_id = $1 AND is_mention = true", sessionID).Scan(&stats.GroupMentions)
	if err != nil {
		return nil, err
	}

	// Last Active
	var lastActive sql.NullTime
	err = r.DB.QueryRow("SELECT MAX(timestamp) FROM messages_log WHERE session_id = $1", sessionID).Scan(&lastActive)
	if err == nil && lastActive.Valid {
		stats.LastActive = &lastActive.Time
	}

	// Daily Stats (Last 7 days)
	rows, err := r.DB.Query(`
		SELECT to_char(timestamp, 'YYYY-MM-DD') as date, COUNT(*)
		FROM messages_log
		WHERE session_id = $1 AND timestamp > NOW() - INTERVAL '7 days'
		GROUP BY date
		ORDER BY date ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ds model.DailyStat
		if err := rows.Scan(&ds.Date, &ds.Count); err == nil {
			stats.DailyStats = append(stats.DailyStats, ds)
		}
	}

	return stats, nil
}

func (r *AnalyticsRepository) GetUniqueContacts(sessionID string) ([]model.Contact, error) {
	query := `
		SELECT from_number, MAX(timestamp) as last_active, COUNT(*) as message_count
		FROM messages_log
		WHERE session_id = $1 AND direction = 'incoming'
		GROUP BY from_number
		ORDER BY last_active DESC
	`
	rows, err := r.DB.Query(query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []model.Contact
	for rows.Next() {
		var c model.Contact
		if err := rows.Scan(&c.PhoneNumber, &c.LastActive, &c.MessageCount); err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}
	return contacts, nil
}
