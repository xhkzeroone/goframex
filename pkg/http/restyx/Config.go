package restyx

import (
	"time"
)

type Config struct {
	Url        string            `mapstructure:"url" yaml:"url"`
	Timeout    time.Duration     `mapstructure:"timeout" yaml:"timeout"`
	RetryCount int               `mapstructure:"retry_count" yaml:"retry_count"`
	RetryWait  time.Duration     `mapstructure:"retry_wait" yaml:"retry_wait"`
	Headers    map[string]string `mapstructure:"headers" yaml:"headers"`
	Debug      bool              `mapstructure:"debug" yaml:"debug"`
}
