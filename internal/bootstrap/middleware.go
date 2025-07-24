package bootstrap

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.io/xhkzeroone/goframex/pkg/http/ginx"
	"github.io/xhkzeroone/goframex/pkg/logger/logrusx"
)

// LoggingMiddleware - Middleware cơ bản không cần dependencies
func LoggingMiddleware() ginx.Middleware {
	return func(next ginx.HandlerFunc) ginx.HandlerFunc {
		return func(ctx *ginx.Context) error {
			start := time.Now()

			// Log request
			logrusx.Log.Infof("Request: %s %s", ctx.Method(), ctx.Path())

			err := next(ctx)

			// Log response
			duration := time.Since(start)
			logrusx.Log.Infof("Response: %d, Duration: %v", ctx.Status(), duration)

			return err
		}
	}
}

// AuthMiddleware - Middleware cần truy cập database để validate token
func AuthMiddleware(container *MiddlewareContainer) ginx.Middleware {
	return func(next ginx.HandlerFunc) ginx.HandlerFunc {
		return func(ctx *ginx.Context) error {
			token := ctx.Headers()["Authorization"]
			if token == "" {
				ctx.JSON(401, gin.H{"error": "Authorization header required"})
				return nil
			}

			// Có thể gọi database để validate token
			// user, err := container.Repositories.UserRepository.GetByToken(token)
			// if err != nil {
			//     ctx.JSON(401, gin.H{"error": "Invalid token"})
			//     return nil
			// }

			// Set user info vào context
			// ctx.Set("user", user)

			return next(ctx)
		}
	}
}

// RateLimitMiddleware - Middleware cần truy cập cache để rate limiting
func RateLimitMiddleware(container *MiddlewareContainer) ginx.Middleware {
	return func(next ginx.HandlerFunc) ginx.HandlerFunc {
		return func(ctx *ginx.Context) error {
			clientIP := ctx.Headers()["X-Forwarded-For"]
			if clientIP == "" {
				clientIP = ctx.Headers()["X-Real-IP"]
			}

			// Có thể sử dụng cache để check rate limit
			// key := fmt.Sprintf("rate_limit:%s", clientIP)
			// count, err := container.Infrastructure.Cache.Get(key)
			// if err != nil {
			//     // Handle error
			// }

			// if count > 100 { // 100 requests per minute
			//     ctx.JSON(429, gin.H{"error": "Rate limit exceeded"})
			//     return nil
			// }

			return next(ctx)
		}
	}
}

// ValidationMiddleware - Middleware validate request body
func ValidationMiddleware() ginx.Middleware {
	return func(next ginx.HandlerFunc) ginx.HandlerFunc {
		return func(ctx *ginx.Context) error {
			// Validate request body nếu cần
			if len(ctx.Body()) > 0 {
				// Có thể validate JSON schema hoặc struct tags
			}

			return next(ctx)
		}
	}
}

// PermissionMiddleware - Middleware check permission với database
func PermissionMiddleware(container *MiddlewareContainer, permission string) ginx.Middleware {
	return func(next ginx.HandlerFunc) ginx.HandlerFunc {
		return func(ctx *ginx.Context) error {
			// user := ctx.Get("user")
			// if user == nil {
			//     ctx.JSON(401, gin.H{"error": "User not authenticated"})
			//     return nil
			// }

			// Check permission từ database
			// hasPermission, err := container.Repositories.UserRepository.HasPermission(user.ID, permission)
			// if err != nil {
			//     ctx.JSON(500, gin.H{"error": "Permission check failed"})
			//     return nil
			// }

			// if !hasPermission {
			//     ctx.JSON(403, gin.H{"error": "Insufficient permissions"})
			//     return nil
			// }

			return next(ctx)
		}
	}
}

// AdminAuthMiddleware - Middleware cho admin routes
func AdminAuthMiddleware(container *MiddlewareContainer) ginx.Middleware {
	return func(next ginx.HandlerFunc) ginx.HandlerFunc {
		return func(ctx *ginx.Context) error {
			// Check admin role
			// user := ctx.Get("user")
			// if user == nil {
			//     ctx.JSON(401, gin.H{"error": "Admin authentication required"})
			//     return nil
			// }

			// isAdmin, err := container.Repositories.UserRepository.IsAdmin(user.ID)
			// if err != nil || !isAdmin {
			//     ctx.JSON(403, gin.H{"error": "Admin access required"})
			//     return nil
			// }

			return next(ctx)
		}
	}
}

// DatabaseTransactionMiddleware - Middleware để wrap request trong database transaction
func DatabaseTransactionMiddleware(container *MiddlewareContainer) ginx.Middleware {
	return func(next ginx.HandlerFunc) ginx.HandlerFunc {
		return func(ctx *ginx.Context) error {
			// Bắt đầu transaction
			// tx := container.Infrastructure.DB.Begin()
			// if tx.Error != nil {
			//     return tx.Error
			// }

			// Set transaction vào context
			// ctx.Set("db_tx", tx)

			err := next(ctx)

			// Commit hoặc rollback transaction
			// if err != nil {
			//     tx.Rollback()
			// } else {
			//     tx.Commit()
			// }

			return err
		}
	}
}

// ExternalServiceMiddleware - Middleware để gọi external service
func ExternalServiceMiddleware(container *MiddlewareContainer, serviceName string) ginx.Middleware {
	return func(next ginx.HandlerFunc) ginx.HandlerFunc {
		return func(ctx *ginx.Context) error {
			// Có thể gọi external service để validate hoặc enrich data
			// switch serviceName {
			// case "user_service":
			//     userData, err := container.ExternalServices.UserService.GetUserInfo(ctx.Headers()["User-Id"])
			//     if err != nil {
			//         ctx.JSON(500, gin.H{"error": "External service unavailable"})
			//         return nil
			//     }
			//     ctx.Set("user_data", userData)
			// }

			return next(ctx)
		}
	}
}

// CacheMiddleware - Middleware để cache response
func CacheMiddleware(container *MiddlewareContainer, ttl time.Duration) ginx.Middleware {
	return func(next ginx.HandlerFunc) ginx.HandlerFunc {
		return func(ctx *ginx.Context) error {
			// Tạo cache key từ request
			_ = fmt.Sprintf("cache:%s:%s", ctx.Method(), ctx.Path())

			// Check cache trước
			// cached, err := container.Infrastructure.Cache.Get(cacheKey)
			// if err == nil && cached != nil {
			//     ctx.JSON(200, cached)
			//     return nil
			// }

			err := next(ctx)

			// Cache response nếu thành công
			// if err == nil && ctx.Status() == 200 {
			//     container.Infrastructure.Cache.Set(cacheKey, ctx.Response(), ttl)
			// }

			return err
		}
	}
}
