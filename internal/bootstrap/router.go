package bootstrap

import (
	"github.io/xhkzeroone/goframex/pkg/http/ginx"
)

// MiddlewareContainer chứa các dependencies cần thiết cho middleware
type MiddlewareContainer struct {
	Infrastructure   *Infrastructure
	Repositories     *Repositories
	ExternalServices *ExternalServices
}

// RegisterRoutes với middleware support
func RegisterRoutes(server *ginx.Server, handlers *Handlers, middlewareContainer *MiddlewareContainer) {

	// Global middleware
	server.Use(LoggingMiddleware())
	server.Use(AuthMiddleware(middlewareContainer))
	server.Use(RateLimitMiddleware(middlewareContainer))

	// User routes với middleware riêng
	userGroup := server.Group("/users", PermissionMiddleware(middlewareContainer, "user:access"))
	userGroup.POST("", handlers.CreateUserHandler, ValidationMiddleware())
	userGroup.GET("/:id", handlers.GetUserByIDHandler)
	userGroup.GET("", handlers.GetUsersHandler)
	userGroup.PUT("/:id", handlers.UpdateUserHandler, ValidationMiddleware())
	userGroup.DELETE("/:id", handlers.DeleteUserHandler)

	// API v1 routes với versioning
	apiV1Group := server.Group("/api/v1", PermissionMiddleware(middlewareContainer, "api:access"))
	apiV1Group.GET("/users", handlers.GetUsersHandler)
}
