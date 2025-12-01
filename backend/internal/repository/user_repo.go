package repository

import (
	"database/sql"
	"errors"
	"time"
	"wago-backend/internal/model"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) CreateUser(pin string) (*model.User, error) {
	var user model.User
	query := `
		INSERT INTO users (pin) 
		VALUES ($1) 
		RETURNING id, pin, created_at, updated_at, last_login`

	err := r.DB.QueryRow(query, pin).Scan(
		&user.ID,
		&user.PIN,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetUserByPIN(pin string) (*model.User, error) {
	var user model.User
	query := `
		SELECT id, pin, created_at, updated_at, last_login 
		FROM users 
		WHERE pin = $1`

	err := r.DB.QueryRow(query, pin).Scan(
		&user.ID,
		&user.PIN,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLogin,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) UpdateLastLogin(userID string) error {
	query := `UPDATE users SET last_login = $1 WHERE id = $2`
	_, err := r.DB.Exec(query, time.Now(), userID)
	return err
}
