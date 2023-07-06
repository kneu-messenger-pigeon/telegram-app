package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestRunApp(t *testing.T) {
	setTestEnvVars := func() {
		loadTestBaseConfigVars()
		_ = os.Setenv("TELEGRAM_TOKEN", "test-token")
		_ = os.Setenv("TELEGRAM_OFFLINE", "1")
	}

	t.Run("Run with mock config", func(t *testing.T) {
		setTestEnvVars()

		var out bytes.Buffer

		running := true
		go func() {
			maxEndTime := time.Now().Add(time.Second)
			for running && maxEndTime.After(time.Now()) &&
				!strings.Contains(out.String(), TelegramControllerStartedMessage) {
				time.Sleep(time.Millisecond * 200)
			}
			time.Sleep(time.Millisecond * 300)
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}()

		err := runApp(&out)
		running = false

		runtime.Gosched()
		runtime.Gosched()
		time.Sleep(time.Millisecond * 500)
		outputString := out.String()
		fmt.Println(outputString)

		assert.NoError(t, err)
		assert.Contains(t, outputString, TelegramControllerStartedMessage)
		assert.Contains(t, outputString, "Started consuming ")
	})

	t.Run("Run with wrong env file", func(t *testing.T) {
		previousWd, err := os.Getwd()
		assert.NoErrorf(t, err, "Failed to get working dir: %s", err)
		tmpDir := os.TempDir() + "/telegram-app-run-dir"
		tmpEnvFilepath := tmpDir + "/.env"

		defer func() {
			_ = os.Chdir(previousWd)
			_ = os.Remove(tmpEnvFilepath)
			_ = os.Remove(tmpDir)
		}()

		if _, err := os.Stat(tmpDir); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(tmpDir, os.ModePerm)
			assert.NoErrorf(t, err, "Failed to create tmp dir %s: %s", tmpDir, err)
		}
		if _, err := os.Stat(tmpEnvFilepath); errors.Is(err, os.ErrNotExist) {
			err := os.Mkdir(tmpEnvFilepath, os.ModePerm)
			assert.NoErrorf(t, err, "Failed to create tmp  %s/.env: %s", tmpDir, err)
		}

		err = os.Chdir(tmpDir)
		assert.NoErrorf(t, err, "Failed to change working dir: %s", err)

		var out bytes.Buffer
		err = runApp(&out)
		assert.Error(t, err, "Expected for error")
		assert.Containsf(
			t, err.Error(), "Error loading .env file",
			"Expected for Load config error, got: %s", err,
		)
	})

	t.Run("Run with wrong redis driver", func(t *testing.T) {
		_ = os.Setenv("REDIS_DSN", "//")
		defer os.Unsetenv("REDIS_DSN")

		var out bytes.Buffer
		err := runApp(&out)

		expectedError := errors.New("redis: invalid URL scheme: ")

		assert.Error(t, err, "Expected for error")
		assert.Equal(t, expectedError, err, "Expected for another error, got %s", err)
	})
}

func TestHandleExitError(t *testing.T) {
	t.Run("Handle exit error", func(t *testing.T) {
		var actualExitCode int
		var out bytes.Buffer

		testCases := map[error]int{
			errors.New("dummy error"): ExitCodeMainError,
			nil:                       0,
		}

		for err, expectedCode := range testCases {
			out.Reset()
			actualExitCode = handleExitError(&out, err)

			assert.Equalf(
				t, expectedCode, actualExitCode,
				"Expect handleExitError(%v) = %d, actual: %d",
				err, expectedCode, actualExitCode,
			)
			if err == nil {
				assert.Empty(t, out.String(), "Error is not empty")
			} else {
				assert.Contains(t, out.String(), err.Error(), "error output hasn't error description")
			}
		}
	})
}
