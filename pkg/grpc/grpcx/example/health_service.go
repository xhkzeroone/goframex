package main

import (
	"context"
	"github.io/xhkzeroone/goframex/pkg/grpc/grpcx/proto/healthpb"
)

type HealthService struct {
	healthpb.UnimplementedHealthServer
}

func NewHealthService() healthpb.HealthServer {
	return &HealthService{}
}

func (s *HealthService) Liveness(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{
		Status: healthpb.HealthCheckResponse_SERVING,
	}, nil
}

func (s *HealthService) Readiness(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{
		Status: healthpb.HealthCheckResponse_SERVING,
	}, nil
}

func (s *HealthService) Terminate(ctx context.Context, req *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	return &healthpb.HealthCheckResponse{
		Status: healthpb.HealthCheckResponse_TERMINATING,
	}, nil
}
