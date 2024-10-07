package infrastructure

import (
	"fmt"

	"github.com/spf13/viper"
)

func LoadConfig(configFile string) (*Config, error) {
	v := viper.New()
	v.SetConfigName(configFile) // name of config file (without extension)
	v.AddConfigPath("./config")
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config file not found: %w", err)
		} else {
			return nil, fmt.Errorf("unable to parse config.file: %w", err)
		}
	}
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}
