package utils

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

// ParseUserIDFromToken validates the JWT and extracts the user_id claim.
func ParseUserIDFromToken(tokenString, secret string) (string, error) {
	if tokenString == "" {
		return "", errors.New("missing token")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}
	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return "", errors.New("invalid user ID in token")
	}
	return userID, nil
}
