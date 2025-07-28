# AppContainer - Dependency Injection Container

## Tổng quan

AppContainer là một dependency injection container được thiết kế để quản lý tất cả các instance trong ứng dụng một cách có tổ chức và dễ bảo trì.

## Cấu trúc

### 1. Infrastructure Layer
```go
type Infrastructure struct {
    DB    *gormx.DataSource
    Cache *redisx.Redis
}
```
- Quản lý các tài nguyên cơ sở hạ tầng như database và cache
- Được khởi tạo đầu tiên vì các layer khác phụ thuộc vào nó

### 2. Repository Layer
```go
type Repositories struct {
    UserRepository    domain.UserRepository
    ProductRepository domain.ProductRepository
}
```
- Chứa các repository interface để tương tác với database
- Sử dụng interface để đảm bảo loose coupling

### 3. External Services Layer
```go
type ExternalServices struct {
    UserService    domain.UserService
    ProductService domain.ProductService
}
```
- Quản lý các external service như validation, external API calls
- Cung cấp business logic cho các service bên ngoài

### 4. Usecase Layer
```go
type Usecases struct {
    UserUsecase    domain.UserUsecase
    ProductUsecase domain.ProductUsecase
}
```
- Chứa business logic chính của ứng dụng
- Kết hợp repository và external service để thực hiện các use case

### 5. Handler Layer
```go
type Handlers struct {
    // User handlers
    CreateUserHandler     *http.CreateUserHandler
    GetUserByIDHandler    *http.GetUserByIDHandler
    GetUsersHandler       *http.GetUsersHandler
    UpdateUserHandler     *http.UpdateUserHandler
    DeleteUserHandler     *http.DeleteUserHandler
    
    // Product handlers
    CreateProductHandler  *http.CreateProductHandler
    GetProductByIDHandler *http.GetProductByIDHandler
    GetProductsHandler    *http.GetProductsHandler
    UpdateProductHandler  *http.UpdateProductHandler
    DeleteProductHandler  *http.DeleteProductHandler
}
```
- Quản lý tất cả HTTP handlers
- Nhóm theo domain (User, Product) để dễ quản lý

## Quy trình khởi tạo

1. **Logger**: Khởi tạo logger đầu tiên
2. **Config**: Load configuration
3. **Infrastructure**: Khởi tạo database và cache
4. **Repositories**: Tạo các repository với dependency injection
5. **External Services**: Khởi tạo các external service
6. **Usecases**: Tạo use cases với repository và service
7. **Handlers**: Tạo handlers với use cases
8. **Server**: Khởi tạo server và register routes

## Lợi ích

### 1. Tổ chức rõ ràng
- Mỗi layer có trách nhiệm riêng biệt
- Dễ dàng tìm và sửa đổi code
- Giảm complexity của từng component

### 2. Dependency Management
- Tất cả dependencies được quản lý tập trung
- Dễ dàng thay đổi implementation
- Testing dễ dàng hơn với mock

### 3. Scalability
- Dễ dàng thêm domain mới
- Có thể mở rộng handlers, repositories, services
- Maintainable khi project lớn

### 4. Type Safety
- Sử dụng interface để đảm bảo type safety
- Compile-time checking cho dependencies
- Giảm runtime errors

## Cách sử dụng

```go
// Khởi tạo container
app, err := bootstrap.NewContainer()
if err != nil {
    log.Fatalf("Failed to initialize usecase: %v", err)
}

// Truy cập các component
db := app.Infrastructure.DB
cache := app.Infrastructure.Cache
userRepo := app.Repositories.UserRepository
userService := app.ExternalServices.UserService
userUsecase := app.Usecases.UserUsecase
createUserHandler := app.Handlers.CreateUserHandler

// Start server
if err := app.Start(); err != nil {
    log.Printf("Server error: %v", err)
}
```

## Thêm domain mới

Để thêm một domain mới (ví dụ: Order), cần:

1. **Thêm vào Repositories**:
```go
type Repositories struct {
    UserRepository    domain.UserRepository
    ProductRepository domain.ProductRepository
    OrderRepository   domain.OrderRepository  // New
}
```

2. **Thêm vào ExternalServices**:
```go
type ExternalServices struct {
    UserService    domain.UserService
    ProductService domain.ProductService
    OrderService   domain.OrderService  // New
}
```

3. **Thêm vào Usecases**:
```go
type Usecases struct {
    UserUsecase    domain.UserUsecase
    ProductUsecase domain.ProductUsecase
    OrderUsecase   domain.OrderUsecase  // New
}
```

4. **Thêm vào Handlers**:
```go
type Handlers struct {
    // ... existing handlers
    CreateOrderHandler  *http.CreateOrderHandler  // New
    GetOrderHandler     *http.GetOrderHandler     // New
    // ... more handlers
}
```

5. **Cập nhật initialization functions**:
```go
func initRepositories(infrastructure *Infrastructure) *Repositories {
    return &Repositories{
        UserRepository:    database.NewUserRepository(infrastructure.DB, infrastructure.Cache),
        ProductRepository: database.NewProductRepository(infrastructure.DB, infrastructure.Cache),
        OrderRepository:   database.NewOrderRepository(infrastructure.DB, infrastructure.Cache), // New
    }
}
```

## Best Practices

1. **Single Responsibility**: Mỗi struct chỉ có một trách nhiệm
2. **Dependency Injection**: Sử dụng interface thay vì concrete types
3. **Loose Coupling**: Các component không phụ thuộc trực tiếp vào nhau
4. **Testability**: Dễ dàng mock và test từng component
5. **Maintainability**: Code dễ đọc và bảo trì 