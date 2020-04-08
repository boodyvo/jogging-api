package auth

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidToken = status.Error(codes.InvalidArgument, "invalid token")
)
