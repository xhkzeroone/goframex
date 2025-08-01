package grpcx

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

func ServerLoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	// Get client info
	var clientIP string
	if p, ok := peer.FromContext(ctx); ok {
		clientIP = p.Addr.String()
	}
	log.Printf("[gRPC] --> %s | From: %s | Request: %+v", info.FullMethod, clientIP, req)
	// Call actual handler
	resp, err := handler(ctx, req)
	st, _ := status.FromError(err)
	log.Printf("[gRPC] <-- %s | Status: %s | Duration: %s | Response: %+v", info.FullMethod, st.Code(), time.Since(start), resp)
	return resp, err
}
