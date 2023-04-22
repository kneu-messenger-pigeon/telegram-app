package main

import (
	"context"
	"fmt"
	"github.com/kneu-messenger-pigeon/authorizer-client"
	"github.com/kneu-messenger-pigeon/score-client"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"io"
	"os"
	"time"

	tele "gopkg.in/telebot.v3"
)

const ExitCodeMainError = 1

const clientName = "telegram-app"

func main() {
	os.Exit(handleExitError(os.Stderr, runApp(os.Stdout)))
}

func runApp(out io.Writer) error {
	var opt *redis.Options
	var bot *tele.Bot
	var redisClient *redis.Client

	envFilename := ""
	if _, err := os.Stat(".env"); err == nil {
		envFilename = ".env"
	}

	config, err := loadConfig(envFilename)
	if err == nil {
		opt, err = redis.ParseURL(config.redisDsn)
	}

	if err == nil {
		redisClient = redis.NewClient(opt)
		_, err = redisClient.Ping(context.Background()).Result()
	}

	if err != nil {
		_, _ = fmt.Fprintf(out, "Failed to connect to redisClient: %s\n", err.Error())
	}

	pref := tele.Settings{
		Token:   config.telegramToken,
		Offline: config.telegramOffline,
		URL:     config.telegramURL,
		Poller: &tele.LongPoller{
			Timeout: time.Second * 30,
		},
		ParseMode: tele.ModeHTML,
	}

	if err == nil {
		bot, err = tele.NewBot(pref)
	}

	if err != nil {
		return err
	}

	userRepository := &UserRepository{
		out:   out,
		redis: redisClient,
	}

	kafkaDialer := &kafka.Dialer{
		Timeout:   config.kafkaTimeout,
		DualStack: kafka.DefaultDialer.DualStack,
	}

	userAuthorizedEventHandler := &UserAuthorizedEventHandler{
		repository: userRepository,
		clientName: clientName,
	}

	userLogoutHandler := UserLogoutHandler{
		out:    out,
		Client: clientName,
		writer: &kafka.Writer{
			Addr:     kafka.TCP(config.kafkaHost),
			Topic:    "authorized_users",
			Balancer: &kafka.LeastBytes{},
		},
	}

	userAuthorizedEventProcessor := KafkaConsumerProcessor{
		out: out,
		reader: kafka.NewReader(
			kafka.ReaderConfig{
				Brokers:     []string{config.kafkaHost},
				GroupID:     clientName,
				Topic:       "authorized_users",
				MinBytes:    10,
				MaxBytes:    10e3,
				MaxWait:     time.Second,
				MaxAttempts: config.kafkaAttempts,
				Dialer:      kafkaDialer,
			},
		),
		handler:         userAuthorizedEventHandler,
		commitThreshold: defaultCommitThreshold,
	}

	scoreChangedEventHandler := &ScoreChangedEventHandler{}

	scoreChangedEventProcessor := KafkaConsumerProcessor{
		out: out,
		reader: kafka.NewReader(
			kafka.ReaderConfig{
				Brokers:     []string{config.kafkaHost},
				GroupID:     clientName,
				Topic:       "scores_changes_feed",
				MinBytes:    10,
				MaxBytes:    10e3,
				MaxWait:     time.Second,
				MaxAttempts: config.kafkaAttempts,
				Dialer:      kafkaDialer,
			},
		),
		handler:         &ScoreChangedEventHandler{},
		commitThreshold: defaultCommitThreshold,
	}

	telegramController := TelegramController{
		out: out,
		bot: bot,
		authorizerClient: &authorizer.Client{
			Host:       config.authorizerHost,
			Secret:     config.appSecret,
			ClientName: clientName,
		},
		userRepository:    userRepository,
		userLogoutHandler: userLogoutHandler,
		scoreClient: &score.Client{
			Host: config.scoreStorageApiHost,
		},
		userAuthorizedEventQueue: userAuthorizedEventHandler.GetEventQueue(),
		scoreChangedEventQueue:   scoreChangedEventHandler.GetEventQueue(),
	}

	executor := ExecutorLoop{
		out: nil,
		executorPool: [ExecutorLoopPoolSize]ExecutorInterface{
			&telegramController,
			&userAuthorizedEventProcessor,
			&scoreChangedEventProcessor,
		},
	}

	defer func() {
		redisClient.Save(context.Background())
		_ = redisClient.Close()
	}()

	executor.execute()

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
