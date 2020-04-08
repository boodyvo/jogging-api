package api

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidInputData  = status.Error(codes.InvalidArgument, "invalid input data")
	ErrInvalidEmail      = status.Error(codes.InvalidArgument, "email is invalid format")
	ErrUserNotFound      = status.Error(codes.NotFound, "user not found")
	ErrUserAlreadyExists = status.Error(codes.InvalidArgument, "user already exists")
	ErrTrackingNotFound  = status.Error(codes.NotFound, "tracking not found")
	ErrTokenNotFound     = status.Error(codes.NotFound, "token not found")
	ErrUnauthorized      = status.Error(codes.Unauthenticated, "cannot parse authorization token")
	ErrForbidden         = status.Error(codes.PermissionDenied, "forbidden")
	ErrInvalidFilter     = status.Error(codes.InvalidArgument, "cannot parse filter")
)
