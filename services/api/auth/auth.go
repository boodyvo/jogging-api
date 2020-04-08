package auth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/google/uuid"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	log "github.com/sirupsen/logrus"

	"github.com/dgrijalva/jwt-go"

	"github.com/boodyvo/jogging-api/services/api/storage"
)

const (
	scheme = "bearer"

	accessTokenExpirationTime  = 10 * time.Minute
	refreshTokenExpirationTime = 7 * 24 * time.Hour
)

type Service interface {
	GenerateToken(ctx context.Context, user *storage.User) (*storage.Token, error)
	VerifyToken(ctx context.Context, accessToken string) (*storage.Claims, error)
	ParseAuthorizationHeader(ctx context.Context) (*storage.Claims, error)
	RefreshToken(ctx context.Context, refreshToken string) (*storage.Token, error)
}

type ServiceImp struct {
	privateKey *ecdsa.PrivateKey
	store      storage.Storage
	logger     *log.Logger
}

func New(privateKey *ecdsa.PrivateKey, store storage.Storage, logger *log.Logger) Service {
	return &ServiceImp{
		privateKey: privateKey,
		store:      store,
		logger:     logger,
	}
}

func (s *ServiceImp) GenerateToken(ctx context.Context, user *storage.User) (*storage.Token, error) {
	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, err
	}
	if err := s.store.SaveToken(token); err != nil {
		return nil, err
	}

	return token, nil
}

func (s *ServiceImp) VerifyToken(ctx context.Context, accessToken string) (*storage.Claims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &storage.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return s.privateKey.Public(), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claim, ok := token.Claims.(*storage.Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claim, nil
}

// TODO(boodyvo): validate signature of refresh token
func (s *ServiceImp) RefreshToken(ctx context.Context, refreshToken string) (*storage.Token, error) {
	tokenOld, err := s.store.GetToken(refreshToken)
	if err != nil {
		return nil, err
	}
	token, err := s.generateToken(tokenOld.UserID)
	if err != nil {
		return nil, err
	}
	if err := s.store.SaveToken(token); err != nil {
		return nil, err
	}

	return token, nil
}

func (s *ServiceImp) generateToken(userID uuid.UUID) (*storage.Token, error) {
	expiresAtAccess := time.Now().Add(accessTokenExpirationTime)
	claimsAccess := &storage.Claims{
		UserID: userID.String(),
		Type:   storage.AccessType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAtAccess.Unix(),
		},
	}
	tokenJWTAccess := jwt.NewWithClaims(jwt.SigningMethodES384, claimsAccess)
	accessToken, err := tokenJWTAccess.SignedString(s.privateKey)
	if err != nil {
		return nil, err
	}

	expiresAtRefresh := time.Now().Add(refreshTokenExpirationTime)
	claimsRefresh := &storage.Claims{
		UserID: userID.String(),
		Type:   storage.RefreshType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAtRefresh.Unix(),
		},
	}
	tokenJWTRefresh := jwt.NewWithClaims(jwt.SigningMethodES384, claimsRefresh)
	refreshToken, err := tokenJWTRefresh.SignedString(s.privateKey)
	if err != nil {
		return nil, err
	}

	token := &storage.Token{
		UserID:    userID,
		Access:    accessToken,
		Refresh:   refreshToken,
		ExpiresAt: expiresAtAccess,
	}

	return token, nil
}

func (s *ServiceImp) ParseAuthorizationHeader(ctx context.Context) (*storage.Claims, error) {
	accessToken, err := grpc_auth.AuthFromMD(ctx, scheme)
	if err != nil {
		return nil, err
	}

	return s.VerifyToken(ctx, accessToken)
}
