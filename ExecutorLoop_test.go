package main

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/mock"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestEventLoopExecute(t *testing.T) {
	t.Run("ExecutorLoop Execute", func(t *testing.T) {
		out := &bytes.Buffer{}
		matchContext := mock.MatchedBy(func(ctx context.Context) bool { return true })
		matchWaitGroup := mock.MatchedBy(func(wg *sync.WaitGroup) bool { wg.Done(); return true })

		connector := NewMockExecutorInterface(t)

		connector.On("Execute", matchContext, matchWaitGroup).Return().Times(ExecutorLoopPoolSize)

		connectorPool := [ExecutorLoopPoolSize]ExecutorInterface{}
		for i := 0; i < ExecutorLoopPoolSize; i++ {
			connectorPool[i] = connector
		}

		eventloop := ExecutorLoop{
			out:          out,
			executorPool: connectorPool,
		}

		go func() {
			time.Sleep(time.Nanosecond * 200)
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		}()
		eventloop.execute()

		connector.AssertExpectations(t)
	})
}
