package ymlx

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func Load(cfg interface{}) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("cfg must be a non-nil pointer to a struct")
	}

	loadFile()

	if err := viper.Unmarshal(cfg); err != nil {
		log.Printf("Can not unmarshal config into struct: %v", err)
		return err
	}
	return nil
}

func resolveEnvInViper() {
	settings := viper.AllSettings()
	for key, value := range settings {
		if strVal, ok := value.(string); ok {
			if envVal, exists := os.LookupEnv(strVal); exists {
				viper.Set(key, envVal)
			}
		}
	}
}

func loadFile() {
	dir, _ := os.Getwd()
	configPath := filepath.Join(dir, "./config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Can not load config.yml: %v", filepath.Join(configPath, "config.yml"))
	}

	env := strings.ToLower(os.Getenv("APP_ENV"))
	if env != "" {
		viper.SetConfigName("config-" + env)
		err := viper.MergeInConfig()
		if err != nil {
			return
		}
	}
	resolveEnvInViper()
}
