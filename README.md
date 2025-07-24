# GoFrameX - Clean Architecture Go Framework

A modern Go web framework built with Clean Architecture principles, featuring domain-driven design, dependency injection, and comprehensive CRUD operations.

## 🏗️ Architecture

```
internal/
├── domain/                 # Domain Layer (Entities & Interfaces)
│   ├── user/              # User Domain
│   └── product/           # Product Domain
├── application/           # Application Layer (Use Cases)
│   ├── user/              # User Business Logic
│   └── product/           # Product Business Logic
├── infrastructure/        # Infrastructure Layer
│   ├── database/          # Database implementations
│   └── external/          # External service implementations
├── interfaces/            # Interface Layer (HTTP handlers)
│   └── http/              # HTTP handlers
│       ├── user/          # User HTTP handlers
│       └── product/       # Product HTTP handlers
└── bootstrap/             # Application bootstrap
```

## 🚀 Features

- **Clean Architecture**: Clear separation of concerns
- **Domain-Driven Design**: Business logic centered around domains
- **Dependency Injection**: Centralized dependency management
- **CRUD Operations**: Full CRUD for User and Product entities
- **Caching**: Redis integration for performance
- **Database**: PostgreSQL with GORM
- **Logging**: Structured logging with Logrus
- **Configuration**: YAML-based configuration
- **HTTP Server**: Gin-based HTTP server with middleware
- **Health Checks**: Built-in health check endpoints

## 📋 Prerequisites

- Go 1.23+
- PostgreSQL 12+
- Redis 6+

## 🛠️ Installation

### 1. Clone the repository
```bash
git clone <repository-url>
cd goframex
```

### 2. Install dependencies
```bash
go mod tidy
```

### 3. Setup Database

#### Option A: Using Docker (Recommended)
```bash
# Start PostgreSQL
docker run --name postgres-goframex \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=goframex \
  -p 5432:5432 \
  -d postgres:15

# Start Redis
docker run --name redis-goframex \
  -p 6379:6379 \
  -d redis:7-alpine
```

#### Option B: Manual Setup
1. Install PostgreSQL and Redis
2. Create database:
```sql
CREATE DATABASE goframex;
```
3. Run the setup script:
```bash
psql -U postgres -f scripts/setup-db.sql
```

### 4. Configure Environment

Update `resources/config.yml` with your database and Redis settings:

```yaml
database:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "password"
  dbname: "goframex"
  sslmode: "disable"
  debug: true

cache:
  host: "localhost"
  port: "6379"
  password: ""
  db: 0
```

### 5. Build and Run
```bash
# Build
go build -o goframex cmd/Main.go

# Run
./goframex
```

Or run directly:
```bash
go run cmd/Main.go
```

## 📡 API Endpoints

### Health Checks
- `GET /api/v1/ping` - Health check
- `GET /api/v1/liveness` - Liveness probe
- `GET /api/v1/readiness` - Readiness probe

### User Management
- `POST /api/v1/users` - Create user
- `GET /api/v1/users/:id` - Get user by ID
- `GET /api/v1/users` - Get all users (with pagination)
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Product Management
- `POST /api/v1/products` - Create product
- `GET /api/v1/products/:id` - Get product by ID
- `GET /api/v1/products` - Get all products (with pagination)
- `GET /api/v1/products?category=Electronics` - Get products by category
- `PUT /api/v1/products/:id` - Update product
- `DELETE /api/v1/products/:id` - Delete product

## 📝 Example Requests

### Create User
```bash
curl -X POST http://localhost:9090/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "password123",
    "age": 30,
    "phone": "+1234567890",
    "address": "123 Main St"
  }'
```

### Create Product
```bash
curl -X POST http://localhost:9090/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "iPhone 15",
    "description": "Latest iPhone model",
    "price": 999.99,
    "stock": 50,
    "category": "Electronics"
  }'
```

### Get Users with Pagination
```bash
curl "http://localhost:9090/api/v1/users?limit=10&offset=0"
```

### Get Products by Category
```bash
curl "http://localhost:9090/api/v1/products?category=Electronics&limit=10&offset=0"
```

## 🔧 Configuration

The application uses YAML configuration located in `resources/config.yml`:

```yaml
server:
  host: "0.0.0.0"
  port: "9090"
  mode: "debug"
  rootPath: "/api/v1"

database:
  host: "localhost"
  port: "5432"
  user: "postgres"
  password: "password"
  dbname: "goframex"
  schema: "public"
  sslmode: "disable"
  debug: true
  driver: "postgres"
  max_open_conns: 10
  max_idle_conns: 5
  conn_max_lifetime: 3600

cache:
  host: "localhost"
  port: "6379"
  password: ""
  db: 0

logger:
  level: "info"
  timestamp_format: "2006-01-02 15:04:05"
  pattern: "%timestamp% | %level% | %requestId% | %file%:%line% | %function% | %message%"
```

## 🏗️ Adding New Domains

To add a new domain (e.g., `order`):

1. **Create Domain Entity**:
```bash
mkdir -p internal/domain/order
touch internal/domain/order/entity.go
```

2. **Create Application Layer**:
```bash
mkdir -p internal/application/order
touch internal/application/order/usecase.go
```

3. **Create Infrastructure Layer**:
```bash
touch internal/infrastructure/database/order_repository.go
touch internal/infrastructure/external/order_service.go
```

4. **Create Interface Layer**:
```bash
mkdir -p internal/interfaces/http/order
touch internal/interfaces/http/order/handler.go
touch internal/interfaces/http/order/request.go
```

5. **Update Bootstrap**:
- Add dependencies in `internal/bootstrap/AppContainer.go`
- Add routes in `internal/bootstrap/router.go`

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...
```

## 📊 Monitoring

The application provides built-in health checks and structured logging:

- **Health Checks**: `/api/v1/ping`, `/api/v1/liveness`, `/api/v1/readiness`
- **Logging**: Structured JSON logging with request tracing
- **Metrics**: Request/response logging with timing information

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License.

## 🆘 Support

For support and questions:
- Create an issue in the repository
- Check the documentation
- Review the example code 