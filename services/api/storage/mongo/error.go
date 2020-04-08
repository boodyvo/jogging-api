package mongo

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrCannotCreateReport = status.Error(codes.Internal, "cannot create report")
)
