package config

import (
	"exam_bot/logger"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	BotToken    string `yaml:"bot_token" env-required:"true"`
	LogPath     string `yaml:"log_path" env-required:"true"`
}

func Load(file string) (Config, error) {
	var cfg Config
	data, err := os.ReadFile(file)
	if err != nil {
		logger.Error().Err(err).Str("file", file).Msg("Failed to read config file")
		return cfg, fmt.Errorf("error reading config file: %v", err)
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		logger.Error().Err(err).Str("file", file).Msg("Failed to parse config file")
		return cfg, fmt.Errorf("error parsing config file: %v", err)
	}

	if cfg.BotToken == "" {
		logger.Error().Msg("bot_token is required in config")
		return cfg, fmt.Errorf("bot_token is required in config")
	}
	if cfg.StoragePath == "" {
		logger.Error().Msg("storage_path is required in config")
		return cfg, fmt.Errorf("storage_path is required in config")
	}
	if cfg.LogPath == "" {
		logger.Error().Msg("log_path is required in config")
		return cfg, fmt.Errorf("log_path is required in config")
	}

	logger.Info().Str("file", file).Msg("Config loaded successfully")
	return cfg, nil
}
