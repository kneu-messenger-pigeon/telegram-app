package main

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
	"time"
)

var testClientName = "test-client"

var expectedBaseConfig = BaseConfig{
	clientName:          testClientName,
	appSecret:           "test_Secret_test123",
	kafkaHost:           "KAFKA:9999",
	kafkaTimeout:        time.Second * 10,
	kafkaAttempts:       0,
	scoreStorageApiHost: "http://localhost:8080",
	authorizerHost:      "http://localhost:8082",
	redisOptions: &redis.Options{
		Network: "tcp",
		Addr:    "REDIS:6379",
	},
}

func TestLoadBaseConfigFromEnvVars(t *testing.T) {
	t.Run("FromEnvVars", func(t *testing.T) {
		_ = os.Unsetenv("KAFKA_TIMEOUT")
		_ = os.Unsetenv("KAFKA_ATTEMPTS")
		_ = os.Setenv("APP_SECRET", expectedBaseConfig.appSecret)
		_ = os.Setenv("KAFKA_HOST", expectedBaseConfig.kafkaHost)
		_ = os.Setenv("KAFKA_TIMEOUT", strconv.Itoa(int(expectedBaseConfig.kafkaTimeout.Seconds())))
		_ = os.Setenv("REDIS_DSN", BuildRedisDsn(expectedBaseConfig.redisOptions))
		_ = os.Setenv("SCORE_STORAGE_API_HOST", expectedBaseConfig.scoreStorageApiHost)
		_ = os.Setenv("AUTHORIZER_HOST", expectedBaseConfig.authorizerHost)

		baseConfig, err := LoadBaseConfig("", testClientName)

		assert.NoErrorf(t, err, "got unexpected error %s", err)
		assertBaseConfig(t, expectedBaseConfig, baseConfig)
		assert.Equalf(t, expectedBaseConfig, baseConfig, "Expected for %v, actual: %v", expectedBaseConfig, baseConfig)
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

		envFileContent += fmt.Sprintf("APP_SECRET=%s\n", expectedBaseConfig.appSecret)
		envFileContent += fmt.Sprintf("KAFKA_HOST=%s\n", expectedBaseConfig.kafkaHost)
		envFileContent += fmt.Sprintf("REDIS_DSN=%s\n", BuildRedisDsn(expectedBaseConfig.redisOptions))
		envFileContent += fmt.Sprintf("SCORE_STORAGE_API_HOST=%s\n", expectedBaseConfig.scoreStorageApiHost)
		envFileContent += fmt.Sprintf("AUTHORIZER_HOST=%s\n", expectedBaseConfig.authorizerHost)

		testEnvFilename := "TestLoadBaseConfigFromFile.env"
		err := os.WriteFile(testEnvFilename, []byte(envFileContent), 0644)
		defer func(name string) {
			_ = os.Remove(name)
		}(testEnvFilename)

		assert.NoErrorf(t, err, "got unexpected while write file %s error %s", testEnvFilename, err)

		baseConfig, err := LoadBaseConfig(testEnvFilename, testClientName)

		assert.NoErrorf(t, err, "got unexpected error %s", err)
		assertBaseConfig(t, expectedBaseConfig, baseConfig)
		assert.Equalf(t, expectedBaseConfig, baseConfig, "Expected for %v, actual: %v", expectedBaseConfig, baseConfig)
	})

	t.Run("empty APP_SECRET", func(t *testing.T) {
		_ = os.Unsetenv("KAFKA_TIMEOUT")
		_ = os.Unsetenv("KAFKA_ATTEMPTS")
		_ = os.Setenv("APP_SECRET", "")
		_ = os.Setenv("KAFKA_HOST", "")
		_ = os.Setenv("REDIS_DSN", "")
		_ = os.Setenv("SCORE_STORAGE_API_HOST", "")
		_ = os.Setenv("AUTHORIZER_HOST", "")

		baseConfig, err := LoadBaseConfig("", testClientName)

		assert.Error(t, err, "LoadBaseConfig() should exit with error, actual error is nil")
		assert.Equalf(
			t, "empty APP_SECRET", err.Error(),
			"Expected for error with empty KAFKA_HOST, actual: %s", err.Error(),
		)

		assert.Emptyf(
			t, baseConfig.appSecret,
			"Expected for empty baseConfig.redisDsn, actual %s", baseConfig.appSecret,
		)

		assert.Emptyf(
			t, baseConfig.kafkaHost,
			"Expected for empty baseConfig.kafkaHost, actual %s", baseConfig.kafkaHost,
		)

		assert.Emptyf(
			t, baseConfig.redisOptions,
			"Expected for empty baseConfig.redisOptions, actual %s", baseConfig.redisOptions,
		)

		assert.Emptyf(
			t, baseConfig.scoreStorageApiHost,
			"Expected for empty baseConfig.scoreStorageApiHost, actual %s", baseConfig.scoreStorageApiHost,
		)
		assert.Emptyf(
			t, baseConfig.authorizerHost,
			"Expected for empty baseConfig.authorizerHost, actual %s", baseConfig.authorizerHost,
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

		config, err := LoadBaseConfig("", testClientName)

		assert.Error(t, err, "LoadBaseConfig() should exit with error, actual error is nil")
		assert.Equalf(
			t, "empty KAFKA_HOST", err.Error(),
			"Expected for error with empty KAFKA_HOST, actual: %s", err.Error(),
		)

		assert.Emptyf(
			t, config.kafkaHost,
			"Expected for empty config.kafkaHost, actual %s", config.kafkaHost,
		)

		assert.Emptyf(
			t, config.redisOptions,
			"Expected for empty config.redisOptions, actual %s", config.redisOptions,
		)

		assert.Emptyf(
			t, config.scoreStorageApiHost,
			"Expected for empty config.scoreStorageApiHost, actual %s", config.scoreStorageApiHost,
		)
		assert.Emptyf(
			t, config.authorizerHost,
			"Expected for empty config.authorizerHost, actual %s", config.authorizerHost,
		)

	})

	t.Run("empty SCORE_STORAGE_API_HOST", func(t *testing.T) {
		_ = os.Setenv("KAFKA_HOST", "dummy-not-empty")
		_ = os.Setenv("REDIS_DSN", "dummy-not-empty")
		_ = os.Setenv("SCORE_STORAGE_API_HOST", "")
		_ = os.Setenv("AUTHORIZER_HOST", "")
		_ = os.Setenv("TELEGRAM_TOKEN", "")

		config, err := LoadBaseConfig("", testClientName)

		assert.Error(t, err, "LoadBaseConfig() should exit with error, actual error is nil")
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
	})

	t.Run("empty REDIS_DSN", func(t *testing.T) {
		_ = os.Unsetenv("KAFKA_TIMEOUT")
		_ = os.Unsetenv("KAFKA_ATTEMPTS")
		_ = os.Setenv("APP_SECRET", "dummy-not-empty")
		_ = os.Setenv("KAFKA_HOST", "dummy-not-empty")
		_ = os.Setenv("SCORE_STORAGE_API_HOST", "dummy-not-empty")
		_ = os.Setenv("AUTHORIZER_HOST", "dummy-not-empty")
		_ = os.Setenv("REDIS_DSN", "")

		config, err := LoadBaseConfig("", testClientName)

		assert.Error(t, err, "LoadBaseConfig() should exit with error, actual error is nil")
		assert.Equalf(
			t, "redis: invalid URL scheme: ", err.Error(),
			"Expected for error with empty KAFKA_HOST, actual: %s", err.Error(),
		)

		assert.Emptyf(
			t, config.redisOptions,
			"Expected for empty config.redisOptions.Addr, actual %s", config.redisOptions,
		)
		assert.Emptyf(
			t, config.scoreStorageApiHost,
			"Expected for empty config.scoreStorageApiHost, actual %s", config.scoreStorageApiHost,
		)
		assert.Emptyf(
			t, config.authorizerHost,
			"Expected for empty config.authorizerHost, actual %s", config.authorizerHost,
		)
	})

	t.Run("empty AUTHORIZER_HOST", func(t *testing.T) {
		_ = os.Setenv("APP_SECRET", "dummy-not-empty")
		_ = os.Setenv("KAFKA_HOST", "dummy-not-empty")
		_ = os.Setenv("REDIS_DSN", "dummy-not-empty")
		_ = os.Setenv("SCORE_STORAGE_API_HOST", "dummy-not-empty")
		_ = os.Setenv("AUTHORIZER_HOST", "")
		_ = os.Setenv("TELEGRAM_TOKEN", "")

		baseConfig, err := LoadBaseConfig("", testClientName)

		assert.Error(t, err, "LoadBaseConfig() should exit with error, actual error is nil")
		assert.Equalf(
			t, "empty AUTHORIZER_HOST", err.Error(),
			"Expected for error with empty KAFKA_HOST, actual: %s", err.Error(),
		)
		assert.Emptyf(
			t, baseConfig.authorizerHost,
			"Expected for empty baseConfig.authorizerHost, actual %s", baseConfig.authorizerHost,
		)
	})

	t.Run("NotExistConfigFile", func(t *testing.T) {
		_ = os.Unsetenv("KAFKA_TIMEOUT")
		_ = os.Unsetenv("KAFKA_ATTEMPTS")
		_ = os.Setenv("REDIS_DSN", "")
		_ = os.Setenv("KAFKA_HOST", "")

		config, err := LoadBaseConfig("not-exists.env", testClientName)

		assert.Error(t, err, "LoadBaseConfig() should exit with error, actual error is nil")
		assert.Equalf(
			t, "Error loading not-exists.env file: open not-exists.env: no such file or directory", err.Error(),
			"Expected for not exist file error, actual: %s", err.Error(),
		)
		assert.Emptyf(
			t, config.redisOptions,
			"Expected for empty config.redisOptions, actual %s", config.redisOptions,
		)
		assert.Emptyf(
			t, config.scoreStorageApiHost,
			"Expected for empty config.scoreStorageApiHost, actual %s", config.scoreStorageApiHost,
		)
	})
}

func assertBaseConfig(t *testing.T, expected BaseConfig, actual BaseConfig) {
	assert.Equal(t, expected.clientName, actual.clientName)
	assert.Equal(t, expected.appSecret, actual.appSecret)
	assert.Equal(t, expected.kafkaHost, actual.kafkaHost)
	assert.Equal(t, expected.kafkaTimeout, actual.kafkaTimeout)
	assert.Equal(t, expected.kafkaAttempts, actual.kafkaAttempts)
	assert.Equal(t, expected.authorizerHost, actual.authorizerHost)
	assert.Equal(t, expected.redisOptions, actual.redisOptions)
	assert.Equal(t, expected.scoreStorageApiHost, actual.scoreStorageApiHost)
}

func BuildRedisDsn(options *redis.Options) string {
	return "redis://@" + options.Addr + "/" + strconv.Itoa(options.DB)
}
