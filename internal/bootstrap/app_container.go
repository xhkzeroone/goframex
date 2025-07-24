package bootstrap

import (
	"fmt"

	uc "github.io/xhkzeroone/goframex/internal/application"
	"github.io/xhkzeroone/goframex/internal/domain"
	"github.io/xhkzeroone/goframex/internal/infrastructure/database"
	"github.io/xhkzeroone/goframex/internal/infrastructure/external"
	"github.io/xhkzeroone/goframex/internal/interfaces/http"
	"github.io/xhkzeroone/goframex/pkg/cache/redisx"
	"github.io/xhkzeroone/goframex/pkg/database/gormx"
	"github.io/xhkzeroone/goframex/pkg/http/ginx"
	"github.io/xhkzeroone/goframex/pkg/http/restyx"
	"github.io/xhkzeroone/goframex/pkg/logger/logrusx"
	"gorm.io/driver/postgres"
)

// Infrastructure layer
type Infrastructure struct {
	DB         *gormx.DataSource
	Cache      *redisx.Redis
	UserClient *restyx.Client
}

// Repository layer
type Repositories struct {
	UserRepository domain.UserRepository
}

// External services layer
type ExternalServices struct {
	UserService domain.UserService
}

// Usecase layer
type Usecases struct {
	UserUsecase domain.UserUsecase
}

// Handler layer
type Handlers struct {
	// User handlers
	CreateUserHandler  *http.CreateUserHandler
	GetUserByIDHandler *http.GetUserByIDHandler
	GetUsersHandler    *http.GetUsersHandler
	UpdateUserHandler  *http.UpdateUserHandler
	DeleteUserHandler  *http.DeleteUserHandler
}

type AppContainer struct {
	Config           *Config
	Infrastructure   *Infrastructure
	Repositories     *Repositories
	ExternalServices *ExternalServices
	Usecases         *Usecases
	Handlers         *Handlers
	Server           *ginx.Server
}

func (app *AppContainer) Start() error {
	return app.Server.Start()
}

func NewContainer() (*AppContainer, error) {
	// Initialize logger
	if err := logrusx.New(); err != nil {
		return nil, err
	}

	// Initialize config
	config, err := NewConfig()
	if err != nil {
		return nil, err
	}

	// Initialize infrastructure
	infrastructure, err := initInfrastructure(config)
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
	server := initServer(config, handlers, infrastructure, repositories, externalServices)

	return &AppContainer{
		Config:           config,
		Infrastructure:   infrastructure,
		Repositories:     repositories,
		ExternalServices: externalServices,
		Usecases:         usecases,
		Handlers:         handlers,
		Server:           server,
	}, nil
}

func initInfrastructure(config *Config) (*Infrastructure, error) {
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
		CreateUserHandler:  http.NewCreateUserHandler(usecases.UserUsecase),
		GetUserByIDHandler: http.NewGetUserByIDHandler(usecases.UserUsecase),
		GetUsersHandler:    http.NewGetUsersHandler(usecases.UserUsecase),
		UpdateUserHandler:  http.NewUpdateUserHandler(usecases.UserUsecase),
		DeleteUserHandler:  http.NewDeleteUserHandler(usecases.UserUsecase),
	}
}

func initServer(config *Config, handlers *Handlers, infrastructure *Infrastructure, repositories *Repositories, externalServices *ExternalServices) *ginx.Server {
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
