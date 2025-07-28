# GoFrameX - A Modern Go Web Framework

GoFrameX is a robust, modular, and scalable web framework for Go applications that follows clean architecture principles and provides essential tools for building enterprise-grade applications.

## Features

- 🏗️ **Clean Architecture**: Built following domain-driven design and clean architecture principles
- 🔌 **Modular Design**: Easy to extend and customize with a pluggable component system
- 🚀 **High Performance**: Optimized for performance using modern Go practices
- 🔒 **Security First**: Built-in security features and middleware
- 🔄 **Database Integration**: Seamless integration with GORM for database operations
- 📦 **Cache Support**: Redis integration for caching
- 📝 **Structured Logging**: Advanced logging with Logrus
- ⚙️ **Configuration Management**: Flexible configuration using Viper
- 🌐 **HTTP Client**: Built-in HTTP client using Resty
- ⏰ **Task Scheduling**: Cron job support for scheduled tasks

## Project Structure

```plaintext
├── cmd/                    # Application entry points
│   └── Main.go            # Main application
├── internal/              # Private application code
│   ├── application/       # Application business rules
│   ├── bootstrap/         # Application bootstrapping
│   ├── domain/           # Enterprise business rules
│   ├── infrastructure/   # External interfaces (DB, Cache, etc.)
│   └── interfaces/       # Delivery mechanisms (HTTP, gRPC, etc.)
└── pkg/                  # Public libraries
    ├── cache/            # Cache utilities
    ├── config/           # Configuration utilities
    ├── database/         # Database utilities
    ├── http/             # HTTP utilities
    ├── logger/           # Logging utilities
    └── scheduler/        # Scheduling utilities
```

## Dependencies

- Go 1.23.1 or higher
- Gin Web Framework
- GORM - Go ORM
- Redis
- PostgreSQL
- Logrus
- Viper
- Resty
- Cron

## Getting Started

1. Clone the repository:

```bash
git clone https://github.com/xhkzeroone/goframex.git
```

2. Install dependencies:

```bash
go mod download
```

3. Configure your application:

- Copy `resources/config.yml` and adjust settings as needed

4. Run the application:

```bash
go run cmd/Main.go
```

## Configuration

The application uses `config.yml` for configuration. You can find the configuration file in the `resources` directory.

## Modules

### HTTP Server (pkg/http/ginx)

- Built on top of Gin framework
- Configurable middleware
- Request/Response handling

### Database (pkg/database/gormx)

- GORM-based database operations
- Connection pooling
- Migration support

### Cache (pkg/cache/redisx)

- Redis integration
- Configurable caching strategies

### Logger (pkg/logger/logrusx)

- Structured logging
- Log levels
- JSON formatting
- Sensitive data masking

### Scheduler (pkg/scheduler/cronx)

- Cron job scheduling
- Background task management

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Author

xhkzeroone
