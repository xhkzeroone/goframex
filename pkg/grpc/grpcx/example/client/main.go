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

func main() {
	cfg := &grpcx.ClientConfig{
		Target: "localhost:50051",
		Debug:  true,
		ClientInterceptors: []grpc.UnaryClientInterceptor{
			grpcx.ClientLoggingInterceptor,
		},
	}

	client, err := grpcx.NewClient(cfg)
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
