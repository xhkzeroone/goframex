package bootstrap

import (
	"context"
	"fmt"
	"github.io/xhkzeroone/goframex/internal/config"
	"github.io/xhkzeroone/goframex/internal/delivery/http"
	"github.io/xhkzeroone/goframex/internal/domain"
	"github.io/xhkzeroone/goframex/internal/infrastructure/database"
	"github.io/xhkzeroone/goframex/internal/infrastructure/external"
	uc "github.io/xhkzeroone/goframex/internal/usecase"
	"github.io/xhkzeroone/goframex/pkg/cache/redisx"
	"github.io/xhkzeroone/goframex/pkg/database/gormx"
	"github.io/xhkzeroone/goframex/pkg/http/ginx"
	"github.io/xhkzeroone/goframex/pkg/http/restyx"
	"github.io/xhkzeroone/goframex/pkg/logger/logrusx"
	"gorm.io/driver/postgres"
)

type Infrastructure struct {
	DB         *gormx.DataSource
	Cache      *redisx.Redis
	UserClient *restyx.Client
}

type Repositories struct {
	UserRepository domain.UserRepository
}

type ExternalServices struct {
	UserService domain.UserService
}

type Usecases struct {
	UserUsecase domain.UserUsecase
}

type Handlers struct {
	UserHandler *http.UserHandler
}

type Application struct {
	Config           *config.Config
	Infrastructure   *Infrastructure
	Repositories     *Repositories
	ExternalServices *ExternalServices
	Usecases         *Usecases
	Handlers         *Handlers
	Server           *ginx.Server
}

func (app *Application) Start() error {
	return app.Server.Start()
}

func (app *Application) Stop() error {
	return app.Server.Stop(context.Background())
}

func NewApp() (*Application, error) {
	if err := logrusx.New(); err != nil {
		return nil, err
	}
	logrusx.Log.SetFormatter(&logrusx.JSONFormatter{
		TimestampFormat:       "2006-01-02 15:04:05",
		MsgFormatter:          logrusx.GetMessageFormater(),
		FunctionNameFormatter: logrusx.GetFunctionNameFormatter(),
	})

	cfg, err := config.NewConfig()
	if err != nil {
		return nil, err
	}

	// Initialize infrastructure
	infrastructure, err := initInfrastructure(cfg)
	if err != nil {
		return nil, err
	}

	// Initialize repositories
	repositories := initRepositories(infrastructure)

	// Initialize external services
	externalServices := initExternalServices(infrastructure)

	// Initialize usecases
	usecases := initUsecases(repositories, externalServices)

	// Initialize handlers
	handlers := initHandlers(usecases)

	// Initialize server
	server := initServer(cfg, handlers, infrastructure, repositories, externalServices)

	return &Application{
		Config:           cfg,
		Infrastructure:   infrastructure,
		Repositories:     repositories,
		ExternalServices: externalServices,
		Usecases:         usecases,
		Handlers:         handlers,
		Server:           server,
	}, nil
}

func initInfrastructure(config *config.Config) (*Infrastructure, error) {
	// Initialize database
	db, err := initDatabase(config.Database)
	if err != nil {
		return nil, err
	}

	// Initialize cache
	cache, err := initCache(config.Cache)
	if err != nil {
		return nil, err
	}

	// Initialize external service userClient
	userClient := restyx.New(config.External.UserClient)

	return &Infrastructure{
		DB:         db,
		Cache:      cache,
		UserClient: userClient,
	}, nil
}

func initRepositories(infrastructure *Infrastructure) *Repositories {
	return &Repositories{
		UserRepository: database.NewUserRepository(infrastructure.DB, infrastructure.Cache),
	}
}

func initExternalServices(infrastructure *Infrastructure) *ExternalServices {
	return &ExternalServices{
		UserService: external.NewUserService(infrastructure.UserClient),
	}
}

func initUsecases(repositories *Repositories, externalServices *ExternalServices) *Usecases {
	return &Usecases{
		UserUsecase: uc.NewUserUsecase(repositories.UserRepository, externalServices.UserService),
	}
}

func initHandlers(usecases *Usecases) *Handlers {
	return &Handlers{
		// User handlers
		UserHandler: http.NewUserHandler(usecases.UserUsecase),
	}
}

func initServer(config *config.Config, handlers *Handlers, infrastructure *Infrastructure, repositories *Repositories, externalServices *ExternalServices) *ginx.Server {
	server := ginx.New(config.Server)

	// Health check
	server.HealthCheck()

	// Create middleware container
	middlewareContainer := &MiddlewareContainer{
		Infrastructure:   infrastructure,
		Repositories:     repositories,
		ExternalServices: externalServices,
	}

	// Register routes with middleware support
	RegisterRoutes(server, handlers, middlewareContainer)

	return server
}

func initDatabase(cfg *gormx.Config) (*gormx.DataSource, error) {
	// Build PostgreSQL DSN
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	// Open database with PostgreSQL driver
	db, err := gormx.Open(cfg, gormx.WithDialector(postgres.Open(dsn)))
	if err != nil {
		return nil, err
	}

	logrusx.Log.Info("Database initialized successfully")
	return db, nil
}

func initCache(cfg *redisx.Config) (*redisx.Redis, error) {
	cache, err := redisx.New(cfg)
	if err != nil {
		return nil, err
	}

	logrusx.Log.Info("Cache initialized successfully")
	return cache, nil
}
