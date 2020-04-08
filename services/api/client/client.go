package client

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	pb "github.com/boodyvo/jogging-api/proto/pb/api"
)

// New returns a grpc client to the api service
func New(ctx context.Context, url string) (pb.APIServiceClient, error) {
	connCtx, _ := context.WithTimeout(ctx, 10*time.Second)
	conn, err := grpc.DialContext(
		connCtx,
		url,
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		if err := conn.Close(); err != nil {
			grpclog.Infof("Failed to close connection: %v", err)
		}
	}()

	return pb.NewAPIServiceClient(conn), nil
}
