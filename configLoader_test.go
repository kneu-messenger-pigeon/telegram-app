package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var expectedConfig = Config{
	telegramToken:   "telegram-token",
	telegramOffline: true,
	telegramURL:     "https://api.telegram.org",
}

func loadTestBaseConfigVars() {
	_ = os.Unsetenv("KAFKA_TIMEOUT")
	_ = os.Unsetenv("KAFKA_ATTEMPTS")
	_ = os.Setenv("APP_SECRET", "test-test")
	_ = os.Setenv("KAFKA_HOST", "localhost:29092")
	_ = os.Setenv("REDIS_DSN", "redis://@localhost:6400/2")
	_ = os.Setenv("SCORE_STORAGE_API_HOST", "http://localhost:8083")
	_ = os.Setenv("AUTHORIZER_HOST", "http://localhost:9080")
}

func TestLoadConfigFromEnvVars(t *testing.T) {
	t.Run("FromEnvVars", func(t *testing.T) {
		loadTestBaseConfigVars()
		_ = os.Setenv("TELEGRAM_TOKEN", expectedConfig.telegramToken)
		_ = os.Setenv("TELEGRAM_OFFLINE", "true")
		_ = os.Setenv("TELEGRAM_URL", expectedConfig.telegramURL)

		actualConfig, err := loadConfig("")

		assert.NoError(t, err)
		assertConfig(t, expectedConfig, actualConfig)
	})

	t.Run("empty TELEGRAM_TOKEN", func(t *testing.T) {
		loadTestBaseConfigVars()
		_ = os.Setenv("TELEGRAM_TOKEN", "")

		config, err := loadConfig("")

		assert.Error(t, err, "loadConfig() should exit with error, actual error is nil")
		assert.Equalf(
			t, "empty TELEGRAM_TOKEN", err.Error(),
			"Expected for error with empty TELEGRAM_TOKEN, actual: %s", err.Error(),
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
