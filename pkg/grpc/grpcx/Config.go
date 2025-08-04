package grpcx

type ServerConfig struct {
	Network string `mapstructure:"network" yaml:"network"` // "tcp" hoặc "unix"
	Address string `mapstructure:"address" yaml:"address"` // ":50051" hoặc "/tmp/app.sock"
	Debug   bool   `mapstructure:"debug" yaml:"debug"`
}

type ClientConfig struct {
	Target string `mapstructure:"target" yaml:"target"`
	Debug  bool   `mapstructure:"debug" yaml:"debug"`
	IsTLS  bool   `mapstructure:"is_tls" yaml:"is_tls"`
}

type Config struct {
	Server ServerConfig `mapstructure:"server" yaml:"server"`
	Client ClientConfig `mapstructure:"client" yaml:"client"`
}
