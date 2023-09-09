package main

import (
	"fmt"
	framework "github.com/kneu-messenger-pigeon/client-framework"
	tele "gopkg.in/telebot.v3"
	"io"
	"os"
	"time"
)

const ExitCodeMainError = 1

const clientName = "telegram-app"

func runApp(out io.Writer) error {
	var bot *tele.Bot

	envFilename := ""
	if _, err := os.Stat(".env"); err == nil {
		envFilename = ".env"
	}

	config, err := loadConfig(envFilename)

	pref := tele.Settings{
		Token:   config.telegramToken,
		Offline: config.telegramOffline,
		URL:     config.telegramURL,
		Poller: &tele.LongPoller{
			Timeout: time.Second * 30,
		},
		ParseMode: tele.ModeMarkdownV2,
	}

	if err == nil {
		bot, err = tele.NewBot(pref)
	}

	if err != nil {
		return err
	}

	serviceContainer := framework.NewServiceContainer(config.BaseConfig, out)

	telegramController := &TelegramController{
		out:               out,
		debugLogger:       serviceContainer.DebugLogger,
		bot:               bot,
		composer:          framework.NewMessageComposer(framework.MessageComposerConfig{}),
		userRepository:    serviceContainer.UserRepository,
		userLogoutHandler: serviceContainer.UserLogoutHandler,
		authorizerClient:  serviceContainer.AuthorizerClient,
		scoreClient:       serviceContainer.ScoreClient,
	}

	serviceContainer.SetController(telegramController)

	serviceContainer.Executor.Execute()

	return nil
}

func handleExitError(errStream io.Writer, err error) int {
	if err != nil {
		_, _ = fmt.Fprintln(errStream, err)
	}

	if err != nil {
		return ExitCodeMainError
	}
	return 0
}
