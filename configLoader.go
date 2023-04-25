package main

import (
	"errors"
	"os"
	"strings"
)

type Config struct {
	BaseConfig
	telegramToken   string
	telegramOffline bool
	// for test purpose override with mock server
	telegramURL string
}

func loadConfig(envFilename string) (Config, error) {
	baseConfig, err := LoadBaseConfig(envFilename, clientName)

	if err != nil {
		return Config{}, err
	}

	config := Config{
		BaseConfig:      baseConfig,
		telegramToken:   os.Getenv("TELEGRAM_TOKEN"),
		telegramOffline: os.Getenv("TELEGRAM_OFFLINE") == "1" || strings.ToLower(os.Getenv("TELEGRAM_OFFLINE")) == "true",
		telegramURL:     os.Getenv("TELEGRAM_URL"),
	}

	if config.telegramToken == "" {
		return Config{}, errors.New("empty TELEGRAM_TOKEN")
	}

	return config, nil
}
