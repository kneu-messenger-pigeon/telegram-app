package main

import (
	"fmt"
	framework "github.com/kneu-messenger-pigeon/client-framework"
	tele "gopkg.in/telebot.v3"
	"io"
	"log"
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
		OnError:   TelegramOnError,
	}

	if err == nil {
		bot, err = tele.NewBot(pref)
	}

	if err != nil {
		return err
	}

	serviceContainer := framework.NewServiceContainer(config.BaseConfig, out)
	telegramController := NewTelegramController(serviceContainer, bot, out)
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

func TelegramOnError(err error, c tele.Context) {
	if c != nil {
		studentId := uint32(0)
		if c.Get(contextStudentKey) != nil {
			studentId = getStudent(c).Id
		}

		log.Println(studentId, c.Update().ID, err)
		OnUpdateErrorCount.Inc()
	} else {
		log.Println(err)
		OnErrorCount.Inc()
	}
}
