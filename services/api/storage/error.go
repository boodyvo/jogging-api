package storage

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidCursor = status.Error(codes.NotFound, "invalid cursor")
	ErrNotFound      = status.Error(codes.NotFound, "not found")
	ErrUnknownRole   = status.Error(codes.NotFound, "unknown role")
	ErrUnknownScope  = status.Error(codes.NotFound, "unknown scope")
	ErrUnknownAction = status.Error(codes.NotFound, "unknown action")
)
