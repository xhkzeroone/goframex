package main

import (
	"errors"
	"github.io/xhkzeroone/goframex/pkg/http/ginx"
	"github.io/xhkzeroone/goframex/pkg/logger/logrusx"
	"net/http"
)

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
	srv.GET("/hello", func(ctx *ginx.Context) error {
		ctx.JSON(http.StatusOK, map[string]string{"message": "Hello, world!"})
		return nil
	}) // 200 OK
	srv.GET("/empty", func(ctx *ginx.Context) error {
		return nil
	}) // 204 No Content
	srv.GET("/fail", func(ctx *ginx.Context) error {
		return errors.New("something went wrong")
	}) // 500 Internal Server Error
	srv.POST("/usr", func(ctx *ginx.Context) error {
		name := string(ctx.Body())
		if name == "" {
			ctx.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
			return nil
		}

		ctx.JSON(http.StatusCreated, map[string]string{
			"message": "User created",
			"name":    name,
		})
		return nil
	}) // 201 Created or 400 Bad Request

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
	userGroup.POST("/create/:name", func(ctx *ginx.Context) error {
		name := string(ctx.Body())
		if name == "" {
			ctx.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
			return nil
		}

		ctx.JSON(http.StatusCreated, map[string]string{
			"message": "User created",
			"name":    name,
		})
		return nil
	})

	// Health check endpoints
	srv.HealthCheck()

	// Start ginx
	if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panic(err)
	}
}
