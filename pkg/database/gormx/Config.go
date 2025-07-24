package gormx

type Config struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Port     string `mapstructure:"port" yaml:"port"`
	User     string `mapstructure:"user" yaml:"user"`
	Password string `mapstructure:"password" yaml:"password"`
	DBName   string `mapstructure:"dbname" yaml:"dbname"`
	Schema   string `mapstructure:"schema" yaml:"schema"`
	SSLMode  string `mapstructure:"sslmode" yaml:"sslmode"`
	Debug    bool   `mapstructure:"debug" yaml:"debug"`
	Driver   string `mapstructure:"driver" yaml:"driver"`

	MaxOpenConns    int   `mapstructure:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int   `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime int64 `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime"`
}
