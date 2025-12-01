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
		INSERT INTO sessions (user_id, session_name, webhook_url, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	err := r.DB.QueryRow(
		query,
		session.UserID,
		session.SessionName,
		session.WebhookURL,
		session.Status,
	).Scan(&session.ID, &session.CreatedAt, &session.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (r *SessionRepository) GetSessionsByUserID(userID string) ([]*model.Session, error) {
	query := `
		SELECT id, session_name, webhook_url, status, phone_number, last_connected, created_at, updated_at
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
		SELECT id, user_id, session_name, webhook_url, status, phone_number, device_info, last_connected, created_at, updated_at
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
		SET session_name = $1, webhook_url = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3 AND user_id = $4`

	_, err := r.DB.Exec(query, session.SessionName, session.WebhookURL, session.ID, session.UserID)
	return err
}

func (r *SessionRepository) UpdateSessionStatus(id string, status model.SessionStatus, phoneNumber string, deviceInfo *model.DeviceInfo) error {
	query := `
		UPDATE sessions
		SET status = $1, phone_number = $2, device_info = $3, updated_at = CURRENT_TIMESTAMP, last_connected = CASE WHEN $1 = 'connected' THEN CURRENT_TIMESTAMP ELSE last_connected END
		WHERE id = $4`

	_, err := r.DB.Exec(query, status, phoneNumber, deviceInfo, id)
	return err
}

func (r *SessionRepository) DeleteSession(id string, userID string) error {
	query := `DELETE FROM sessions WHERE id = $1 AND user_id = $2`
	_, err := r.DB.Exec(query, id, userID)
	return err
}
