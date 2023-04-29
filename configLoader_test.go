package main

import (
	framework "github.com/kneu-messenger-pigeon/client-framework"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var expectedConfig = Config{
	BaseConfig:      framework.BaseConfig{},
	telegramToken:   "telegram-token",
	telegramOffline: true,
	telegramURL:     "https://api.telegram.org",
}

func TestLoadConfigFromEnvVars(t *testing.T) {
	t.Run("FromEnvVars", func(t *testing.T) {
		_ = os.Setenv("TELEGRAM_TOKEN", expectedConfig.telegramToken)
		_ = os.Setenv("TELEGRAM_OFFLINE", "true")
		_ = os.Setenv("TELEGRAM_URL", expectedConfig.telegramURL)

		config, err := loadConfig("")

		assert.Error(t, err)
		assert.Equal(t, "empty APP_SECRET", err.Error())
		assertConfig(t, expectedConfig, config)
		assert.Equalf(t, expectedConfig, config, "Expected for %v, actual: %v", expectedConfig, config)
	})

	t.Run("empty TELEGRAM_TOKEN", func(t *testing.T) {
		_ = os.Setenv("TELEGRAM_TOKEN", "")

		config, err := loadConfig("")

		assert.Error(t, err, "loadConfig() should exit with error, actual error is nil")
		assert.Equalf(
			t, "empty TELEGRAM_TOKEN", err.Error(),
			"Expected for error with empty KAFKA_HOST, actual: %s", err.Error(),
		)

		assert.Emptyf(
			t, config.telegramToken,
			"Expected for empty config.telegramToken, actual %s", config.telegramToken,
		)
	})
}

func assertConfig(t *testing.T, expected Config, actual Config) {
	assert.Equal(t, expected.telegramToken, actual.telegramToken)
	assert.Equal(t, expected.telegramOffline, actual.telegramOffline)
	assert.Equal(t, expected.telegramURL, actual.telegramURL)
}
