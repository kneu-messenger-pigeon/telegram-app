package main

import (
	"errors"
	framework "github.com/kneu-messenger-pigeon/client-framework"
	"os"
	"strings"
)

type Config struct {
	framework.BaseConfig
	telegramToken   string
	telegramOffline bool
	// for test purpose override with mock server
	telegramURL string
}

func loadConfig(envFilename string) (Config, error) {
	baseConfig, err := framework.LoadBaseConfig(envFilename, clientName)

	config := Config{
		BaseConfig:      baseConfig,
		telegramToken:   os.Getenv("TELEGRAM_TOKEN"),
		telegramOffline: os.Getenv("TELEGRAM_OFFLINE") == "1" || strings.ToLower(os.Getenv("TELEGRAM_OFFLINE")) == "true",
		telegramURL:     os.Getenv("TELEGRAM_URL"),
	}

	if config.telegramToken == "" && err == nil {
		err = errors.New("empty TELEGRAM_TOKEN")
	}

	return config, err
}
