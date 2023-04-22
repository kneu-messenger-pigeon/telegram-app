package main

import (
	"context"
	"sync"
)

type ExecutorInterface interface {
	Execute(ctx context.Context, wg *sync.WaitGroup)
}
