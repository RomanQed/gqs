package internal

import (
	"context"
	"log/slog"
	"sync"
)

type WorkHandler[T any] func(context.Context, T)

type WorkerPool[T any] struct {
	concurrency int
	queue       int
	wg          sync.WaitGroup
	in          chan T
	ctx         context.Context
	cancel      context.CancelFunc
	log         *slog.Logger
}

func NewWorkerPool[T any](concurrency int, queue int, log *slog.Logger) *WorkerPool[T] {
	return &WorkerPool[T]{
		concurrency: concurrency,
		queue:       queue,
		log:         log,
	}
}

func (wp *WorkerPool[T]) safeHandle(ctx context.Context, wh WorkHandler[T], t T) {
	defer func() {
		if r := recover(); r != nil {
			wp.log.Error("worker panic recovered", "err", r)
		}
	}()
	wh(ctx, t)
}

func (wp *WorkerPool[T]) worker(ctx context.Context, wh WorkHandler[T]) {
	defer wp.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-wp.in:
			wp.safeHandle(ctx, wh, t)
		}
	}
}

func (wp *WorkerPool[T]) Push(t T) bool {
	select {
	case <-wp.ctx.Done():
		return false
	case wp.in <- t:
		return true
	}
}

func (wp *WorkerPool[T]) Start(ctx context.Context, wh WorkHandler[T]) {
	wp.ctx, wp.cancel = context.WithCancel(ctx)
	wp.in = make(chan T, wp.queue)
	for i := 0; i < wp.concurrency; i++ {
		wp.wg.Add(1)
		go wp.worker(wp.ctx, wh)
	}
}

func (wp *WorkerPool[T]) Stop() DoneChan {
	wp.cancel()
	return wrapWaitGroup(&wp.wg)
}
