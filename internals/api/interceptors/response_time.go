package interceptors

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func ResponseTimeIntercepter(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// record the start time
	start := time.Now()

	//call  the handler to proceed with the actual  request
	resp, err := handler(ctx, req)

	//calculate the duration
	duration := time.Since(start)

	// log the request details with the duration
	st, _ := status.FromError(err)

	log.Printf("Method: %s, Status: %s, Duration: %v\n", info.FullMethod, st.Code(), duration)

	// setting metadata
	md := metadata.Pairs("X-Response-Time", duration.String())
	grpc.SetHeader(ctx, md)

	return resp, err
}
