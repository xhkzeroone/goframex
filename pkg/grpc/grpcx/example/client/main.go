package main

import (
	"context"
	"fmt"
	"github.io/xhkzeroone/goframex/pkg/grpc/grpcx"
	"github.io/xhkzeroone/goframex/pkg/grpc/grpcx/proto/healthpb"
	"google.golang.org/grpc"
	"log"
	"time"
)

func AuthInterceptor(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	log.Printf("[gRPC REQUEST] - AuthInterceptor: Method: %s | Request: %+v", method, req)
	err := invoker(ctx, method, req, reply, cc, opts...)
	return err
}

func main() {
	cfg := &grpcx.ClientConfig{
		Target: "localhost:50051",
		Debug:  true,
	}

	client, err := grpcx.NewClient(cfg, grpc.WithChainUnaryInterceptor(grpcx.ClientLoggingInterceptor, AuthInterceptor))
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	stub := grpcx.Stub(client, healthpb.NewHealthClient)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resp, err := stub.Liveness(ctx, &healthpb.HealthCheckRequest{})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	fmt.Println("Response:", resp.Status.String())
}
