package bootstrap

import (
	"github.io/xhkzeroone/goframex/pkg/cache/redisx"
	ymlx "github.io/xhkzeroone/goframex/pkg/config"
	"github.io/xhkzeroone/goframex/pkg/database/gormx"
	"github.io/xhkzeroone/goframex/pkg/http/ginx"
	"github.io/xhkzeroone/goframex/pkg/http/restyx"
	"github.io/xhkzeroone/goframex/pkg/logger/logrusx"
)

type Config struct {
	Server   *ginx.Config    `mapstructure:"server" yaml:"server"`
	Database *gormx.Config   `mapstructure:"database" yaml:"database"`
	Cache    *redisx.Config  `mapstructure:"cache" yaml:"cache"`
	Logger   *logrusx.Config `mapstructure:"logger" yaml:"logger"`
	External *External       `mapstructure:"external" yaml:"external"`
}

type External struct {
	UserClient *restyx.Config `mapstructure:"user-client" yaml:"user-client"`
}

func NewConfig() (*Config, error) {
	config := &Config{}
	if err := ymlx.Load(config); err != nil {
		return nil, err
	}
	return config, nil
}
