package gormx

import (
	"fmt"
	"gorm.io/gorm"
	"log"
	"time"
)

type Option func(o *options)

type options struct {
	dialector  gorm.Dialector
	gormConfig *gorm.Config
	debug      *bool
}

type DataSource struct {
	*gorm.DB
	Config *Config
}

func WithDialector(d gorm.Dialector) Option {
	return func(o *options) {
		o.dialector = d
	}
}

func WithGormConfig(cfg *gorm.Config) Option {
	return func(o *options) {
		o.gormConfig = cfg
	}
}

func WithDebug(debug bool) Option {
	return func(o *options) {
		o.debug = &debug
	}
}

func Open(cfg *Config, opts ...Option) (*DataSource, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config must not be nil")
	}
	opt := &options{}
	for _, o := range opts {
		o(opt)
	}

	gormCfg := &gorm.Config{}
	if opt.gormConfig != nil {
		gormCfg = opt.gormConfig
	}

	db, err := gorm.Open(opt.dialector, gormCfg)
	if err != nil {
		log.Printf("failed to connect database: %v", err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}

	debugMode := cfg.Debug
	if opt.debug != nil {
		debugMode = *opt.debug
	}
	if debugMode {
		db = db.Debug()
		log.Println("GORM debug mode is enabled")
	}

	log.Println("Successfully connected to database")
	return &DataSource{DB: db, Config: cfg}, nil
}

func (p *DataSource) Close() error {
	if p == nil || p.DB == nil {
		return nil
	}
	sqlDB, err := p.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
