package grpcx

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type Registrar struct {
	RegisterFunc       func(server *grpc.Server)
	ServiceDesc        *grpc.ServiceDesc
	Interceptor        []grpc.UnaryServerInterceptor
	MethodInterceptors map[string][]grpc.UnaryServerInterceptor
}

func NewRegistrar[T any](registerFunc func(grpc.ServiceRegistrar, T), desc *grpc.ServiceDesc, impl T, interceptor ...grpc.UnaryServerInterceptor) Registrar {
	return Registrar{
		RegisterFunc: func(s *grpc.Server) {
			registerFunc(s, impl)
		},
		ServiceDesc: desc,
		Interceptor: interceptor,
	}
}

type Server struct {
	*grpc.Server
	listener           net.Listener
	config             *ServerConfig
	interceptors       map[string]grpc.UnaryServerInterceptor   // by service
	methodInterceptors map[string][]grpc.UnaryServerInterceptor // by full method
	globalInterceptors []grpc.UnaryServerInterceptor
}

func NewServer(config *ServerConfig) *Server {
	network := config.Network
	if network == "" {
		network = "tcp" // fallback mặc định
	}

	address := config.Address
	if address == "" {
		address = ":50051" // fallback mặc định
	}

	listener, err := net.Listen(network, address)
	if err != nil {
		log.Fatalf("❌ Failed to listen: %v", err)
	}

	s := &Server{
		config:             config,
		listener:           listener,
		interceptors:       make(map[string]grpc.UnaryServerInterceptor),
		methodInterceptors: make(map[string][]grpc.UnaryServerInterceptor),
	}

	s.Server = grpc.NewServer(
		grpc.UnaryInterceptor(s.dispatchInterceptor()),
	)

	return s
}

func (s *Server) Use(interceptors ...grpc.UnaryServerInterceptor) {
	s.globalInterceptors = append(s.globalInterceptors, interceptors...)
}

func (s *Server) Register(services ...Registrar) {
	for _, svc := range services {
		svc.RegisterFunc(s.Server)

		if svc.ServiceDesc != nil && len(svc.Interceptor) > 0 {
			s.interceptors[svc.ServiceDesc.ServiceName] = ChainUnaryInterceptors(svc.Interceptor...)
		}

		for fullMethod, interceptors := range svc.MethodInterceptors {
			s.methodInterceptors[fullMethod] = append(s.methodInterceptors[fullMethod], interceptors...)
		}

	}
}

func (s *Server) Start() error {
	go s.listenForShutdown()
	log.Printf("🚀 gRPC server is running at %s", s.listener.Addr())
	s.PrintRegisteredServices()
	return s.Serve(s.listener)
}

func (s *Server) PrintRegisteredServices() {
	services := s.GetServiceInfo()
	s.debugLog("📋 Registered gRPC services and methods:")
	for serviceName, info := range services {
		s.debugLog("  • %s", serviceName)
		for _, method := range info.Methods {
			s.debugLog("    └─ %s", method.Name)
		}
	}
}

func (s *Server) GracefulStop() {
	s.debugLog("🧹 Gracefully stopping gRPC server...")
	s.Server.GracefulStop()
}

func (s *Server) StopImmediately() {
	s.debugLog("❌ Force stopping gRPC server...")
	s.Server.Stop()
}

func (s *Server) Shutdown(force bool) {
	if force {
		s.StopImmediately()
	} else {
		s.GracefulStop()
	}
}

func (s *Server) dispatchInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var interceptors []grpc.UnaryServerInterceptor

		// Global interceptors
		if len(s.globalInterceptors) > 0 {
			interceptors = append(interceptors, s.globalInterceptors...)
		}

		// Service-level interceptor
		for serviceName, interceptor := range s.interceptors {
			if strings.HasPrefix(info.FullMethod, "/"+serviceName+"/") {
				interceptors = append(interceptors, interceptor)
				break
			}
		}

		// Method-level interceptors
		if mInterceptors, ok := s.methodInterceptors[info.FullMethod]; ok {
			interceptors = append(interceptors, mInterceptors...)
		}

		if len(interceptors) == 0 {
			return handler(ctx, req)
		}

		chain := ChainUnaryInterceptors(interceptors...)
		return chain(ctx, req, info, handler)
	}
}

func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, finalHandler grpc.UnaryHandler) (interface{}, error) {
		chain := finalHandler
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			chain = func(next grpc.UnaryHandler) grpc.UnaryHandler {
				return func(ctx context.Context, req interface{}) (interface{}, error) {
					return interceptor(ctx, req, info, next)
				}
			}(chain)
		}
		return chain(ctx, req)
	}
}

func (s *Server) listenForShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	sig := <-c
	s.debugLog("📦 Received signal: %s", sig)
	s.GracefulStop()
}

func (s *Server) debugLog(format string, args ...interface{}) {
	if s.config.Debug {
		log.Printf(format, args...)
	}
}
