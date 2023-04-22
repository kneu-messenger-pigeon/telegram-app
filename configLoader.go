package main

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	appSecret           string
	kafkaHost           string
	kafkaTimeout        time.Duration
	kafkaAttempts       int
	scoreStorageApiHost string
	authorizerHost      string
	redisDsn            string
	telegramToken       string
	telegramOffline     bool
	telegramURL         string // for test purpose override with mock server

}

func loadConfig(envFilename string) (Config, error) {
	if envFilename != "" {
		err := godotenv.Load(envFilename)
		if err != nil {
			return Config{}, errors.New(fmt.Sprintf("Error loading %s file: %s", envFilename, err))
		}
	}

	kafkaTimeout, err := strconv.Atoi(os.Getenv("KAFKA_TIMEOUT"))
	if kafkaTimeout == 0 || err != nil {
		kafkaTimeout = 10
	}

	kafkaAttempts, err := strconv.Atoi(os.Getenv("KAFKA_ATTEMPTS"))
	if kafkaAttempts == 0 || err != nil {
		kafkaAttempts = 0
	}

	config := Config{
		appSecret:           os.Getenv("APP_SECRET"),
		kafkaHost:           os.Getenv("KAFKA_HOST"),
		kafkaTimeout:        time.Second * time.Duration(kafkaTimeout),
		kafkaAttempts:       kafkaAttempts,
		redisDsn:            os.Getenv("REDIS_DSN"),
		scoreStorageApiHost: os.Getenv("SCORE_STORAGE_API_HOST"),
		authorizerHost:      os.Getenv("AUTHORIZER_HOST"),
		telegramToken:       os.Getenv("TELEGRAM_TOKEN"),
		telegramOffline:     os.Getenv("TELEGRAM_OFFLINE") == "1" || strings.ToLower(os.Getenv("TELEGRAM_OFFLINE")) == "true",
		telegramURL:         os.Getenv("TELEGRAM_URL"),
	}

	if config.appSecret == "" {
		return Config{}, errors.New("empty APP_SECRET")
	}

	if config.kafkaHost == "" {
		return Config{}, errors.New("empty KAFKA_HOST")
	}

	if config.redisDsn == "" {
		return Config{}, errors.New("empty REDIS_DSN")
	}

	if config.scoreStorageApiHost == "" {
		return Config{}, errors.New("empty SCORE_STORAGE_API_HOST")
	}

	if config.authorizerHost == "" {
		return Config{}, errors.New("empty AUTHORIZER_HOST")
	}

	if config.telegramToken == "" {
		return Config{}, errors.New("empty TELEGRAM_TOKEN")
	}

	return config, nil
}
