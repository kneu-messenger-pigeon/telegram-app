package main

import (
	"github.com/kneu-messenger-pigeon/authorizer-client"
	"github.com/kneu-messenger-pigeon/score-client"
	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"io"
	"time"
)

type ServiceContainer struct {
	UserRepository               *UserRepository
	UserLogoutHandler            *UserLogoutHandler
	AuthorizerClient             *authorizer.Client
	ScoreClient                  *score.Client
	UserAuthorizedEventProcessor *KafkaConsumerProcessor
	ScoreChangedEventProcessor   *KafkaConsumerProcessor
	Executor                     *Executor
	ClientController             ClientControllerInterface
}

func NewServiceContainer(config BaseConfig, out io.Writer) *ServiceContainer {
	redisClient := redis.NewClient(config.redisOptions)

	kafkaDialer := &kafka.Dialer{
		Timeout:   config.kafkaTimeout,
		DualStack: kafka.DefaultDialer.DualStack,
	}

	container := &ServiceContainer{}

	container.UserRepository = &UserRepository{
		out:   out,
		redis: redisClient,
	}

	container.UserLogoutHandler = &UserLogoutHandler{
		out:    out,
		Client: config.clientName,
		writer: &kafka.Writer{
			Addr:     kafka.TCP(config.kafkaHost),
			Topic:    "authorized_users",
			Balancer: &kafka.LeastBytes{},
		},
	}

	container.AuthorizerClient = &authorizer.Client{
		Host:       config.authorizerHost,
		Secret:     config.appSecret,
		ClientName: config.clientName,
	}

	container.ScoreClient = &score.Client{
		Host: config.scoreStorageApiHost,
	}

	container.UserAuthorizedEventProcessor = &KafkaConsumerProcessor{
		out: out,
		handler: &UserAuthorizedEventHandler{
			out:              out,
			clientName:       config.clientName,
			repository:       container.UserRepository,
			serviceContainer: container,
		},
		reader: kafka.NewReader(
			kafka.ReaderConfig{
				Brokers:     []string{config.kafkaHost},
				GroupID:     config.clientName,
				Topic:       "authorized_users",
				MinBytes:    10,
				MaxBytes:    10e3,
				MaxWait:     time.Second,
				MaxAttempts: config.kafkaAttempts,
				Dialer:      kafkaDialer,
			},
		),
	}

	container.ScoreChangedEventProcessor = &KafkaConsumerProcessor{
		out: out,
		handler: &ScoreChangedEventHandler{
			out:              out,
			serviceContainer: container,
		},
		reader: kafka.NewReader(
			kafka.ReaderConfig{
				Brokers:     []string{config.kafkaHost},
				GroupID:     config.clientName,
				Topic:       "scores_changes_feed",
				MinBytes:    10,
				MaxBytes:    10e3,
				MaxWait:     time.Second,
				MaxAttempts: config.kafkaAttempts,
				Dialer:      kafkaDialer,
			},
		),
	}
	container.Executor = &Executor{
		out:              out,
		serviceContainer: container,
	}

	return container
}

func (container *ServiceContainer) SetController(controller ClientControllerInterface) {
	container.ClientController = controller
}
