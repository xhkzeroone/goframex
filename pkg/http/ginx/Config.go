package ginx

type Config struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Port     string `mapstructure:"port" yaml:"port"`
	Mode     string `mapstructure:"mode" yaml:"mode"`
	RootPath string `mapstructure:"rootPath" yaml:"rootPath"`
}

func (c *Config) GetAddr() string {
	return c.Host + ":" + c.Port
}
