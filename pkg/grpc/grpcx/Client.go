package grpcx

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"sync"
	"time"
)

type Client struct {
	conn   *grpc.ClientConn
	config *ClientConfig
	mu     sync.RWMutex
	closed bool
}

func NewClient(cfg *ClientConfig, extraOpts ...grpc.DialOption) (*Client, error) {
	var opts []grpc.DialOption
	var interceptors []grpc.UnaryClientInterceptor

	if cfg.Debug {
		interceptors = append(interceptors, debugUnaryClientInterceptor)
	}

	if !cfg.IsTLS {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if len(cfg.ClientInterceptors) > 0 {
		interceptors = append(interceptors, cfg.ClientInterceptors...)
	}

	if len(interceptors) > 0 {
		opts = append(opts, grpc.WithUnaryInterceptor(chainUnaryClientInterceptors(interceptors...)))
	}

	opts = append(opts, extraOpts...)

	conn, err := grpc.NewClient(cfg.Target, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		config: cfg,
	}, nil
}

func (c *Client) Conn() *grpc.ClientConn {
	return c.conn
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	return c.conn.Close()
}

func Stub[T any](client *Client, fnc func(cc grpc.ClientConnInterface) T) T {
	return fnc(client.conn)
}

func debugUnaryClientInterceptor(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Println("GRPC CLIENT CALL", method, "err:", err, "took:", time.Since(start))
	return err
}

func chainUnaryClientInterceptors(interceptors ...grpc.UnaryClientInterceptor) grpc.UnaryClientInterceptor {
	n := len(interceptors)
	if n == 0 {
		return func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption,
		) error {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
	}
	return func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var chain grpc.UnaryInvoker
		chain = func(currentCtx context.Context, currentMethod string, currentReq interface{}, currentReply interface{}, currentCC *grpc.ClientConn, currentOpts ...grpc.CallOption) error {
			return interceptors[0](currentCtx, currentMethod, currentReq, currentReply, currentCC, buildChain(interceptors[1:], invoker), currentOpts...)
		}
		return chain(ctx, method, req, reply, cc, opts...)
	}
}

func buildChain(interceptors []grpc.UnaryClientInterceptor, finalInvoker grpc.UnaryInvoker) grpc.UnaryInvoker {
	if len(interceptors) == 0 {
		return finalInvoker
	}

	return func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return interceptors[0](ctx, method, req, reply, cc, buildChain(interceptors[1:], finalInvoker), opts...)
	}
}
