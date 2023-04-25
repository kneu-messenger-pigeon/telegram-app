package main

import (
	"context"
	"fmt"
	"io"
	"os/signal"
	"sync"
	"syscall"
)

type Executor struct {
	out              io.Writer
	serviceContainer *ServiceContainer
}

func (executor *Executor) execute() {
	_, err := executor.serviceContainer.UserRepository.redis.Ping(context.Background()).Result()
	if err != nil {
		_, _ = fmt.Fprintf(executor.out, "Failed to connect to redisClient: %s\n", err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	wg := &sync.WaitGroup{}

	wg.Add(3)
	go executor.serviceContainer.ClientController.Execute(ctx, wg)
	go executor.serviceContainer.UserAuthorizedEventProcessor.Execute(ctx, wg)
	go executor.serviceContainer.ScoreChangedEventProcessor.Execute(ctx, wg)

	wg.Wait()

	executor.serviceContainer.UserRepository.redis.Save(context.Background())
}
