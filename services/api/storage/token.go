package storage

import (
	"time"

	"github.com/google/uuid"

	"github.com/dgrijalva/jwt-go"
)

type Token struct {
	UserID    uuid.UUID `json:"user_id" bson:"user_id"`
	Access    string    `json:"access_token" bson:"access_token"`
	Refresh   string    `json:"refresh_token" bson:"refresh_token"`
	ExpiresAt time.Time `json:"expires_at" bson:"expires_at"`
}

type TokenType string

const (
	AccessType  TokenType = "access"
	RefreshType TokenType = "refresh"
)

type Claims struct {
	UserID string    `json:"user_id"`
	Type   TokenType `json:"type"`
	jwt.StandardClaims
}
