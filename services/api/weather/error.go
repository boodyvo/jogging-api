package weather

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrCannotGetWeather = status.Error(codes.Internal, "cannot obtain weather")
)
