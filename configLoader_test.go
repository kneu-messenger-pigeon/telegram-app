package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
	"time"
)

var expectedConfig = Config{
	appSecret:           "test_Secret_test123",
	kafkaHost:           "KAFKA:9999",
	redisDsn:            "REDIS:6379",
	kafkaTimeout:        time.Second * 10,
	kafkaAttempts:       0,
	scoreStorageApiHost: "http://localhost:8080",
	authorizerHost:      "http://localhost:8082",
	telegramToken:       "telegram-token",
	telegramOffline:     true,
	telegramURL:         "https://api.telegram.org",
}

func TestLoadConfigFromEnvVars(t *testing.T) {
	t.Run("FromEnvVars", func(t *testing.T) {
		_ = os.Unsetenv("KAFKA_TIMEOUT")
		_ = os.Unsetenv("KAFKA_ATTEMPTS")
		_ = os.Setenv("APP_SECRET", expectedConfig.appSecret)
		_ = os.Setenv("KAFKA_HOST", expectedConfig.kafkaHost)
		_ = os.Setenv("KAFKA_TIMEOUT", strconv.Itoa(int(expectedConfig.kafkaTimeout.Seconds())))
		_ = os.Setenv("REDIS_DSN", expectedConfig.redisDsn)
		_ = os.Setenv("SCORE_STORAGE_API_HOST", expectedConfig.scoreStorageApiHost)
		_ = os.Setenv("AUTHORIZER_HOST", expectedConfig.authorizerHost)
		_ = os.Setenv("TELEGRAM_TOKEN", expectedConfig.telegramToken)
		_ = os.Setenv("TELEGRAM_OFFLINE", "true")
		_ = os.Setenv("TELEGRAM_URL", expectedConfig.telegramURL)

		config, err := loadConfig("")

		assert.NoErrorf(t, err, "got unexpected error %s", err)
		assertConfig(t, expectedConfig, config)
		assert.Equalf(t, expectedConfig, config, "Expected for %v, actual: %v", expectedConfig, config)
	})

	t.Run("FromFile", func(t *testing.T) {
		var envFileContent string

		_ = os.Unsetenv("APP_SECRET")
		_ = os.Unsetenv("KAFKA_TIMEOUT")
		_ = os.Unsetenv("KAFKA_ATTEMPTS")
		_ = os.Unsetenv("KAFKA_HOST")
		_ = os.Unsetenv("KAFKA_TIMEOUT")
		_ = os.Unsetenv("REDIS_DSN")
		_ = os.Unsetenv("SCORE_STORAGE_API_HOST")
		_ = os.Unsetenv("AUTHORIZER_HOST")
		_ = os.Unsetenv("TELEGRAM_TOKEN")
		_ = os.Unsetenv("TELEGRAM_OFFLINE")
		_ = os.Unsetenv("TELEGRAM_URL")

		envFileContent += fmt.Sprintf("APP_SECRET=%s\n", expectedConfig.appSecret)
		envFileContent += fmt.Sprintf("KAFKA_HOST=%s\n", expectedConfig.kafkaHost)
		envFileContent += fmt.Sprintf("REDIS_DSN=%s\n", expectedConfig.redisDsn)
		envFileContent += fmt.Sprintf("SCORE_STORAGE_API_HOST=%s\n", expectedConfig.scoreStorageApiHost)
		envFileContent += fmt.Sprintf("AUTHORIZER_HOST=%s\n", expectedConfig.authorizerHost)
		envFileContent += fmt.Sprintf("TELEGRAM_TOKEN=%s\n", expectedConfig.telegramToken)
		envFileContent += fmt.Sprintf("TELEGRAM_OFFLINE=%s\n", "true")
		envFileContent += fmt.Sprintf("TELEGRAM_URL=%s\n", expectedConfig.telegramURL)

		testEnvFilename := "TestLoadConfigFromFile.env"
		err := os.WriteFile(testEnvFilename, []byte(envFileContent), 0644)
		defer os.Remove(testEnvFilename)
		assert.NoErrorf(t, err, "got unexpected while write file %s error %s", testEnvFilename, err)

		config, err := loadConfig(testEnvFilename)

		assert.NoErrorf(t, err, "got unexpected error %s", err)
		assertConfig(t, expectedConfig, config)
		assert.Equalf(t, expectedConfig, config, "Expected for %v, actual: %v", expectedConfig, config)
	})

	t.Run("empty APP_SECRET", func(t *testing.T) {
		_ = os.Unsetenv("KAFKA_TIMEOUT")
		_ = os.Unsetenv("KAFKA_ATTEMPTS")
		_ = os.Setenv("APP_SECRET", "")
		_ = os.Setenv("KAFKA_HOST", "")
		_ = os.Setenv("REDIS_DSN", "")
		_ = os.Setenv("SCORE_STORAGE_API_HOST", "")
		_ = os.Setenv("AUTHORIZER_HOST", "")
		_ = os.Setenv("TELEGRAM_TOKEN", "")

		config, err := loadConfig("")

		assert.Error(t, err, "loadConfig() should exit with error, actual error is nil")
		assert.Equalf(
			t, "empty APP_SECRET", err.Error(),
			"Expected for error with empty KAFKA_HOST, actual: %s", err.Error(),
		)

		assert.Emptyf(
			t, config.appSecret,
			"Expected for empty config.redisDsn, actual %s", config.appSecret,
		)

		assert.Emptyf(
			t, config.kafkaHost,
			"Expected for empty config.redisDsn, actual %s", config.redisDsn,
		)

		assert.Emptyf(
			t, config.redisDsn,
			"Expected for empty config.redisDsn, actual %s", config.redisDsn,
		)

		assert.Emptyf(
			t, config.scoreStorageApiHost,
			"Expected for empty config.scoreStorageApiHost, actual %s", config.scoreStorageApiHost,
		)
		assert.Emptyf(
			t, config.authorizerHost,
			"Expected for empty config.authorizerHost, actual %s", config.authorizerHost,
		)

		assert.Emptyf(
			t, config.telegramToken,
			"Expected for empty config.telegramToken, actual %s", config.redisDsn,
		)
	})

	t.Run("empty KAFKA_HOST", func(t *testing.T) {
		_ = os.Unsetenv("KAFKA_TIMEOUT")
		_ = os.Unsetenv("KAFKA_ATTEMPTS")
		_ = os.Setenv("APP_SECRET", "dummy-not-empty")
		_ = os.Setenv("KAFKA_HOST", "")
		_ = os.Setenv("REDIS_DSN", "")
		_ = os.Setenv("SCORE_STORAGE_API_HOST", "")
		_ = os.Setenv("AUTHORIZER_HOST", "")
		_ = os.Setenv("TELEGRAM_TOKEN", "")

		config, err := loadConfig("")

		assert.Error(t, err, "loadConfig() should exit with error, actual error is nil")
		assert.Equalf(
			t, "empty KAFKA_HOST", err.Error(),
			"Expected for error with empty KAFKA_HOST, actual: %s", err.Error(),
		)

		assert.Emptyf(
			t, config.kafkaHost,
			"Expected for empty config.redisDsn, actual %s", config.redisDsn,
		)

		assert.Emptyf(
			t, config.redisDsn,
			"Expected for empty config.redisDsn, actual %s", config.redisDsn,
		)

		assert.Emptyf(
			t, config.scoreStorageApiHost,
			"Expected for empty config.scoreStorageApiHost, actual %s", config.scoreStorageApiHost,
		)
		assert.Emptyf(
			t, config.authorizerHost,
			"Expected for empty config.authorizerHost, actual %s", config.authorizerHost,
		)

		assert.Emptyf(
			t, config.telegramToken,
			"Expected for empty config.telegramToken, actual %s", config.redisDsn,
		)
	})

	t.Run("empty REDIS_DSN", func(t *testing.T) {
		_ = os.Unsetenv("KAFKA_TIMEOUT")
		_ = os.Unsetenv("KAFKA_ATTEMPTS")
		_ = os.Setenv("KAFKA_HOST", "dummy-not-empty")
		_ = os.Setenv("REDIS_DSN", "")
		_ = os.Setenv("SCORE_STORAGE_API_HOST", "")
		_ = os.Setenv("AUTHORIZER_HOST", "")
		_ = os.Setenv("TELEGRAM_TOKEN", "")

		config, err := loadConfig("")

		assert.Error(t, err, "loadConfig() should exit with error, actual error is nil")
		assert.Equalf(
			t, "empty REDIS_DSN", err.Error(),
			"Expected for error with empty KAFKA_HOST, actual: %s", err.Error(),
		)

		assert.Emptyf(
			t, config.redisDsn,
			"Expected for empty config.redisDsn, actual %s", config.redisDsn,
		)
		assert.Emptyf(
			t, config.scoreStorageApiHost,
			"Expected for empty config.scoreStorageApiHost, actual %s", config.scoreStorageApiHost,
		)
		assert.Emptyf(
			t, config.authorizerHost,
			"Expected for empty config.authorizerHost, actual %s", config.authorizerHost,
		)
		assert.Emptyf(
			t, config.telegramToken,
			"Expected for empty config.telegramToken, actual %s", config.redisDsn,
		)
	})

	t.Run("empty SCORE_STORAGE_API_HOST", func(t *testing.T) {
		_ = os.Setenv("KAFKA_HOST", "dummy-not-empty")
		_ = os.Setenv("REDIS_DSN", "dummy-not-empty")
		_ = os.Setenv("SCORE_STORAGE_API_HOST", "")
		_ = os.Setenv("AUTHORIZER_HOST", "")
		_ = os.Setenv("TELEGRAM_TOKEN", "")

		config, err := loadConfig("")

		assert.Error(t, err, "loadConfig() should exit with error, actual error is nil")
		assert.Equalf(
			t, "empty SCORE_STORAGE_API_HOST", err.Error(),
			"Expected for error with empty KAFKA_HOST, actual: %s", err.Error(),
		)

		assert.Emptyf(
			t, config.scoreStorageApiHost,
			"Expected for empty config.scoreStorageApiHost, actual %s", config.scoreStorageApiHost,
		)
		assert.Emptyf(
			t, config.authorizerHost,
			"Expected for empty config.authorizerHost, actual %s", config.authorizerHost,
		)

		assert.Emptyf(
			t, config.telegramToken,
			"Expected for empty config.telegramToken, actual %s", config.redisDsn,
		)
	})

	t.Run("empty AUTHORIZER_HOST", func(t *testing.T) {
		_ = os.Setenv("KAFKA_HOST", "dummy-not-empty")
		_ = os.Setenv("REDIS_DSN", "dummy-not-empty")
		_ = os.Setenv("SCORE_STORAGE_API_HOST", "dummy-not-empty")
		_ = os.Setenv("AUTHORIZER_HOST", "")
		_ = os.Setenv("TELEGRAM_TOKEN", "")

		config, err := loadConfig("")

		assert.Error(t, err, "loadConfig() should exit with error, actual error is nil")
		assert.Equalf(
			t, "empty AUTHORIZER_HOST", err.Error(),
			"Expected for error with empty KAFKA_HOST, actual: %s", err.Error(),
		)

		assert.Emptyf(
			t, config.telegramToken,
			"Expected for empty config.telegramToken, actual %s", config.redisDsn,
		)
	})

	t.Run("empty TELEGRAM_TOKEN", func(t *testing.T) {
		_ = os.Setenv("KAFKA_HOST", "dummy-not-empty")
		_ = os.Setenv("REDIS_DSN", "dummy-not-empty")
		_ = os.Setenv("SCORE_STORAGE_API_HOST", "dummy-not-empty")
		_ = os.Setenv("AUTHORIZER_HOST", "dummy-not-empty")
		_ = os.Setenv("TELEGRAM_TOKEN", "")

		config, err := loadConfig("")

		assert.Error(t, err, "loadConfig() should exit with error, actual error is nil")
		assert.Equalf(
			t, "empty TELEGRAM_TOKEN", err.Error(),
			"Expected for error with empty KAFKA_HOST, actual: %s", err.Error(),
		)

		assert.Emptyf(
			t, config.telegramToken,
			"Expected for empty config.telegramToken, actual %s", config.redisDsn,
		)
	})

	t.Run("NotExistConfigFile", func(t *testing.T) {
		_ = os.Unsetenv("KAFKA_TIMEOUT")
		_ = os.Unsetenv("KAFKA_ATTEMPTS")
		os.Setenv("REDIS_DSN", "")
		os.Setenv("KAFKA_HOST", "")

		config, err := loadConfig("not-exists.env")

		assert.Error(t, err, "loadConfig() should exit with error, actual error is nil")
		assert.Equalf(
			t, "Error loading not-exists.env file: open not-exists.env: no such file or directory", err.Error(),
			"Expected for not exist file error, actual: %s", err.Error(),
		)
		assert.Emptyf(
			t, config.redisDsn,
			"Expected for empty config.redisDsn, actual %s", config.redisDsn,
		)
		assert.Emptyf(
			t, config.scoreStorageApiHost,
			"Expected for empty config.scoreStorageApiHost, actual %s", config.scoreStorageApiHost,
		)
		assert.Emptyf(
			t, config.telegramToken,
			"Expected for empty config.telegramToken, actual %s", config.telegramToken,
		)
	})
}

func assertConfig(t *testing.T, expected Config, actual Config) {
	assert.Equal(t, expected.redisDsn, actual.redisDsn)
	assert.Equal(t, expected.scoreStorageApiHost, actual.scoreStorageApiHost)
	assert.Equal(t, expected.telegramToken, actual.telegramToken)
	assert.Equal(t, expected.telegramOffline, actual.telegramOffline)
}
