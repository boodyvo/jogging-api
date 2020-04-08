package storage

import "github.com/google/uuid"

type Storage interface {
	// User CRUD
	SaveUser(user *User) error
	UpdateUser(user *User) error
	DeleteUser(id uuid.UUID) error
	GetUser(id uuid.UUID) (*User, error)
	GetUserByEmail(email string) (*User, error)
	ListUsers(filter *UserFilter) (*ListUsersResponse, error)

	// Tracking CRUD
	SaveTracking(tracking *Tracking) error
	UpdateTracking(tracking *Tracking) error
	DeleteTracking(id uuid.UUID) error
	GetTracking(id uuid.UUID) (*Tracking, error)
	ListTrackings(filter *TrackingFilter) (*ListTrackingsResponse, error)
	ListTrackingsForUser(filter *TrackingFilter) (*ListTrackingsResponse, error)
	GetReport(filter *ReportFilter) (*Report, error)

	// Token CRUD
	SaveToken(token *Token) error
	DeleteToken(token *Token) error
	GetToken(refreshToken string) (*Token, error)
}
