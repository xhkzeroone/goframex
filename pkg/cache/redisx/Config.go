package redisx

type Config struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Port     string `mapstructure:"port" yaml:"port"`
	Password string `mapstructure:"password" yaml:"password"`
	DB       int    `mapstructure:"gormx" yaml:"gormx"`
}

func (c *Config) GetAddr() string {
	return c.Host + ":" + c.Port
}
