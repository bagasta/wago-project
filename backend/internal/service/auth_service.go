package service

import (
	"errors"
	"time"
	"wago-backend/internal/config"
	"wago-backend/internal/model"
	"wago-backend/internal/repository"
	"wago-backend/internal/utils"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	UserRepo *repository.UserRepository
	Config   *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		UserRepo: userRepo,
		Config:   cfg,
	}
}

func (s *AuthService) GeneratePIN() (*model.User, error) {
	// Generate unique PIN
	var pin string
	var err error

	// Try up to 5 times to generate a unique PIN
	for i := 0; i < 5; i++ {
		pin, err = utils.GeneratePIN(6)
		if err != nil {
			return nil, err
		}

		// Check if PIN exists
		existingUser, err := s.UserRepo.GetUserByPIN(pin)
		if err != nil {
			return nil, err
		}
		if existingUser == nil {
			break
		}
		if i == 4 {
			return nil, errors.New("failed to generate unique PIN")
		}
	}

	return s.UserRepo.CreateUser(pin)
}

func (s *AuthService) Login(pin string) (string, *model.User, error) {
	user, err := s.UserRepo.GetUserByPIN(pin)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		return "", nil, errors.New("invalid credentials")
	}

	// Update last login
	if err := s.UserRepo.UpdateLastLogin(user.ID); err != nil {
		return "", nil, err
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.Config.JWTSecret))
	if err != nil {
		return "", nil, err
	}

	return tokenString, user, nil
}
