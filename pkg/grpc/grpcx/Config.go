package grpcx

type Config struct {
	Network string `mapstructure:"network" yaml:"network"` // "tcp" hoặc "unix"
	Address string `mapstructure:"address" yaml:"address"` // ":50051" hoặc "/tmp/app.sock"
	Debug   bool   `mapstructure:"debug" yaml:"debug"`
}
