package main

import (
	"context"
	"fmt"
	"github.io/xhkzeroone/goframex/pkg/grpc/grpcx"
	"github.io/xhkzeroone/goframex/pkg/grpc/grpcx/proto/healthpb"
	"google.golang.org/grpc"
	"log"
)

func MetricsInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Call actual handler
	fmt.Println("MetricsInterceptor")
	resp, err := handler(ctx, req)
	return resp, err
}

func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		fmt.Println("auth interceptor")
		resp, err := handler(ctx, req)
		return resp, err
	}
}

func main() {
	server := grpcx.NewServer(&grpcx.ServerConfig{
		Network: "tcp",
		Address: ":50051",
		Debug:   true,
	})

	server.Use(grpcx.ServerLoggingInterceptor)

	svc := grpcx.NewRegistrar(healthpb.RegisterHealthServer, &healthpb.Health_ServiceDesc, NewHealthService(), MetricsInterceptor)

	svc.MethodInterceptors = map[string][]grpc.UnaryServerInterceptor{
		"/proto.Health/Liveness": {
			AuthInterceptor(),
		},
	}

	server.Register(svc)

	if err := server.Start(); err != nil {
		log.Fatalf("❌ Server exited: %v", err)
	}
}
