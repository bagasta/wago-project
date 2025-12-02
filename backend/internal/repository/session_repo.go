package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"wago-backend/internal/model"
)

type SessionRepository struct {
	DB *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{DB: db}
}

func (r *SessionRepository) CreateSession(session *model.Session) (*model.Session, error) {
	query := `
		INSERT INTO sessions (user_id, session_name, webhook_url, status, is_group_response_enabled)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := r.DB.QueryRow(
		query,
		session.UserID,
		session.SessionName,
		session.WebhookURL,
		session.Status,
		session.IsGroupResponseEnabled,
	).Scan(&session.ID, &session.CreatedAt, &session.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *SessionRepository) GetSessionsByUserID(userID string) ([]*model.Session, error) {
	query := `
		SELECT id, session_name, webhook_url, status, phone_number, last_connected, is_group_response_enabled, created_at, updated_at
		FROM sessions
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*model.Session
	for rows.Next() {
		var s model.Session
		var lastConnected sql.NullTime
		var phoneNumber sql.NullString

		err := rows.Scan(
			&s.ID,
			&s.SessionName,
			&s.WebhookURL,
			&s.Status,
			&phoneNumber,
			&lastConnected,
			&s.IsGroupResponseEnabled,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if lastConnected.Valid {
			s.LastConnected = &lastConnected.Time
		}
		if phoneNumber.Valid {
			s.PhoneNumber = phoneNumber.String
		}

		sessions = append(sessions, &s)
	}

	return sessions, nil
}

func (r *SessionRepository) GetSessionByID(id string) (*model.Session, error) {
	var s model.Session
	var lastConnected sql.NullTime
	var phoneNumber sql.NullString
	var deviceInfo []byte

	query := `
		SELECT id, user_id, session_name, webhook_url, status, phone_number, device_info, last_connected, is_group_response_enabled, created_at, updated_at
		FROM sessions
		WHERE id = $1`

	err := r.DB.QueryRow(query, id).Scan(
		&s.ID,
		&s.UserID,
		&s.SessionName,
		&s.WebhookURL,
		&s.Status,
		&phoneNumber,
		&deviceInfo,
		&lastConnected,
		&s.IsGroupResponseEnabled,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if lastConnected.Valid {
		s.LastConnected = &lastConnected.Time
	}
	if phoneNumber.Valid {
		s.PhoneNumber = phoneNumber.String
	}
	if deviceInfo != nil {
		// Assuming DeviceInfo implements Scanner, but here we scan into []byte first to be safe or if jsonb is null
		// Actually, let's use the model's Scan method if we passed &s.DeviceInfo directly, but s.DeviceInfo is a pointer.
		// So we need to handle it.
		// For simplicity in this raw query scan, let's just unmarshal if not nil.
		// But wait, I defined Scan on *DeviceInfo.
		// Let's re-scan properly or just unmarshal here.
		s.DeviceInfo = &model.DeviceInfo{}
		if err := json.Unmarshal(deviceInfo, s.DeviceInfo); err != nil {
			// ignore error or log it, maybe it's null
			s.DeviceInfo = nil
		}
	}

	return &s, nil
}

func (r *SessionRepository) UpdateSession(session *model.Session) error {
	query := `
		UPDATE sessions
		SET session_name = $1, webhook_url = $2, is_group_response_enabled = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND user_id = $5`

	_, err := r.DB.Exec(query, session.SessionName, session.WebhookURL, session.IsGroupResponseEnabled, session.ID, session.UserID)
	return err
}

func (r *SessionRepository) UpdateSessionStatus(id string, status model.SessionStatus, phoneNumber string, deviceInfo *model.DeviceInfo) error {
	var query string
	var args []interface{}

	if status == model.SessionStatusConnected {
		query = `
			UPDATE sessions
			SET status = $1,
			    phone_number = $2,
			    device_info = $3,
			    updated_at = CURRENT_TIMESTAMP,
			    last_connected = CURRENT_TIMESTAMP
			WHERE id = $4`
		args = []interface{}{status, phoneNumber, deviceInfo, id}
	} else {
		query = `
			UPDATE sessions
			SET status = $1,
			    phone_number = $2,
			    device_info = $3,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $4`
		args = []interface{}{status, phoneNumber, deviceInfo, id}
	}

	res, err := r.DB.Exec(query, args...)
	if err != nil {
		return err
	}

	if rows, _ := res.RowsAffected(); rows == 0 {
		return errors.New("no session updated (invalid session id)")
	}
	return nil
}

func (r *SessionRepository) DeleteSession(id string, userID string) error {
	query := `DELETE FROM sessions WHERE id = $1 AND user_id = $2`
	_, err := r.DB.Exec(query, id, userID)
	return err
}

func (r *SessionRepository) GetSessionsByStatus(status model.SessionStatus) ([]*model.Session, error) {
	query := `
		SELECT id, user_id, session_name, webhook_url, status, phone_number, device_info, last_connected, is_group_response_enabled, created_at, updated_at
		FROM sessions
		WHERE status = $1`

	rows, err := r.DB.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*model.Session
	for rows.Next() {
		var s model.Session
		var lastConnected sql.NullTime
		var phoneNumber sql.NullString
		var deviceInfo []byte

		err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.SessionName,
			&s.WebhookURL,
			&s.Status,
			&phoneNumber,
			&deviceInfo,
			&lastConnected,
			&s.IsGroupResponseEnabled,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if lastConnected.Valid {
			s.LastConnected = &lastConnected.Time
		}
		if phoneNumber.Valid {
			s.PhoneNumber = phoneNumber.String
		}
		if deviceInfo != nil {
			s.DeviceInfo = &model.DeviceInfo{}
			if err := json.Unmarshal(deviceInfo, s.DeviceInfo); err != nil {
				s.DeviceInfo = nil
			}
		}

		sessions = append(sessions, &s)
	}
	return sessions, nil
}

// GetSessionsWithPhoneNumber returns all sessions that have a stored JID/phone_number.
// This is useful for reconnecting previously paired sessions even if their status
// was not left as "connected" (e.g. after an unexpected restart).
func (r *SessionRepository) GetSessionsWithPhoneNumber() ([]*model.Session, error) {
	query := `
		SELECT id, user_id, session_name, webhook_url, status, phone_number, device_info, last_connected, is_group_response_enabled, created_at, updated_at
		FROM sessions
		WHERE phone_number IS NOT NULL AND phone_number <> ''`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*model.Session
	for rows.Next() {
		var s model.Session
		var lastConnected sql.NullTime
		var phoneNumber sql.NullString
		var deviceInfo []byte

		err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.SessionName,
			&s.WebhookURL,
			&s.Status,
			&phoneNumber,
			&deviceInfo,
			&lastConnected,
			&s.IsGroupResponseEnabled,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if lastConnected.Valid {
			s.LastConnected = &lastConnected.Time
		}
		if phoneNumber.Valid {
			s.PhoneNumber = phoneNumber.String
		}
		if deviceInfo != nil {
			s.DeviceInfo = &model.DeviceInfo{}
			if err := json.Unmarshal(deviceInfo, s.DeviceInfo); err != nil {
				s.DeviceInfo = nil
			}
		}

		sessions = append(sessions, &s)
	}
	return sessions, nil
}
