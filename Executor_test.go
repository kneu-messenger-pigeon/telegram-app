package main

import (
	"context"
	"github.com/stretchr/testify/mock"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestEventLoopExecute(t *testing.T) {
	t.Run("Executor Execute", func(t *testing.T) {
		matchContext := mock.MatchedBy(func(ctx context.Context) bool { return true })
		matchWaitGroup := mock.MatchedBy(func(wg *sync.WaitGroup) bool { wg.Done(); return true })

		connector := NewMockExecutorInterface(t)

		connector.On("Execute", matchContext, matchWaitGroup).Return().Times(ExecutorPoolSize)

		pool := [ExecutorPoolSize]ExecutorInterface{}
		for i := 0; i < ExecutorPoolSize; i++ {
			pool[i] = connector
		}

		executor := Executor{
			executorPool: pool,
		}

		go func() {
			time.Sleep(time.Nanosecond * 200)
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}()
		executor.execute()

		connector.AssertExpectations(t)
	})
}
