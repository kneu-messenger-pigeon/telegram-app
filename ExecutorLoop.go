package main

import (
	"context"
	"io"
	"os/signal"
	"sync"
	"syscall"
)

const ExecutorLoopPoolSize = 3

type ExecutorLoop struct {
	out          io.Writer
	executorPool [ExecutorLoopPoolSize]ExecutorInterface
}

func (eventLoop *ExecutorLoop) execute() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	wg := &sync.WaitGroup{}

	wg.Add(len(eventLoop.executorPool))
	for _, executor := range eventLoop.executorPool {
		go executor.Execute(ctx, wg)
	}

	wg.Wait()
}
