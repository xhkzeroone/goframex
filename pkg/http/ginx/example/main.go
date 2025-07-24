package main

import (
	"errors"
	"github.io/xhkzeroone/goframex/pkg/http/ginx"
	"github.io/xhkzeroone/goframex/pkg/logger/logrusx"
	"net/http"
)

type HelloHandler struct{}

func (h *HelloHandler) Handle(ctx *ginx.Context, headers, query, path map[string]string) error {
	ctx.JSON(http.StatusOK, map[string]string{"message": "Hello, world!"})
	return nil
}

type NoResponseHandler struct{}

func (h *NoResponseHandler) Handle(ctx *ginx.Context, headers, query, path map[string]string) error {
	// Không set response => sẽ trả HTTP 204 No Content
	return nil
}

type ErrorHandler struct{}

func (h *ErrorHandler) Handle(ctx *ginx.Context, headers, query, path map[string]string) error {
	return errors.New("something went wrong")
}

type CreateUserHandler struct{}

func (h *CreateUserHandler) Handle(ctx *ginx.Context, body []byte, headers, query, path map[string]string) error {
	name := string(body)
	if name == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
		return nil
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"message": "User created",
		"name":    name,
	})
	return nil
}

func main() {

	_ = logrusx.New()

	cfg := &ginx.Config{
		Mode:     "debug",
		RootPath: "/api",
		Host:     "127.0.0.1",
		Port:     "8080",
	}

	srv := ginx.New(cfg)

	srv.Use(func(next ginx.HandlerFunc) ginx.HandlerFunc {
		return func(ctx *ginx.Context) error {
			logrusx.Log.Infof("%s %s", ctx.Request.Method, ctx.Request.RequestURI)
			err := next(ctx)
			logrusx.Log.Infof("Response %s", ctx.Response())
			return err
		}
	})
	// Các route cơ bản
	srv.GET("/hello", &HelloHandler{})      // 200 OK
	srv.GET("/empty", &NoResponseHandler{}) // 204 No Content
	srv.GET("/fail", &ErrorHandler{})       // 500 Internal Server Error
	srv.POST("/usr", &CreateUserHandler{})  // 201 Created or 400 Bad Request

	// Group route
	userGroup := srv.Group("/user")
	userGroup.Use(func(next ginx.HandlerFunc) ginx.HandlerFunc {
		return func(ctx *ginx.Context) error {
			logrusx.Log.Infof("/user %s %s", ctx.Request.Method, ctx.Request.RequestURI)
			err := next(ctx)
			logrusx.Log.Infof("/user Response %s", ctx.Response())
			return err
		}
	})
	userGroup.POST("/create/:name", &CreateUserHandler{})

	// Health check endpoints
	srv.HealthCheck()

	// Start ginx
	if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}
