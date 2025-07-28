package bootstrap

import (
	"github.io/xhkzeroone/goframex/pkg/http/ginx"
)

type MiddlewareContainer struct {
	Infrastructure   *Infrastructure
	Repositories     *Repositories
	ExternalServices *ExternalServices
}

func RegisterRoutes(server *ginx.Server, handlers *Handlers, middlewareContainer *MiddlewareContainer) {
	// Global middleware
	server.Use(LoggingMiddleware())
	server.Use(AuthMiddleware(middlewareContainer))
	server.Use(RateLimitMiddleware(middlewareContainer))

	// User routes với middleware riêng
	userGroup := server.Group("/users", PermissionMiddleware(middlewareContainer, "user:access"))
	userGroup.POST("", handlers.UserHandler.CreateUser, ValidationMiddleware())
	userGroup.GET("/:id", handlers.UserHandler.GetUserByID)
	userGroup.GET("", handlers.UserHandler.GetUsers)
	userGroup.PUT("/:id", handlers.UserHandler.UpdateUser, ValidationMiddleware())
	userGroup.DELETE("/:id", handlers.UserHandler.DeleteUser)

	// API v1 routes với versioning
	apiV1Group := server.Group("/api/v1", PermissionMiddleware(middlewareContainer, "api:access"))
	apiV1Group.GET("/users", handlers.UserHandler.GetUsers)
}
