package logrusx

import (
	"encoding/xml"
	"os"
)

type Config struct {
	TimestampFormat string `xml:"timestampFormat"`
	Pattern         string `xml:"pattern"`
	Level           string `xml:"level"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = xml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
