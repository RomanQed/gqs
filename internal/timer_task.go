package internal

import (
	"context"
	"time"
)

type TimerHandler func(context.Context)

type TimerTask struct {
	cancel context.CancelFunc
	done   DoneChan
}

func (t *TimerTask) do(ctx context.Context, h TimerHandler, timeout time.Duration) {
	defer close(t.done)
	ticker := time.NewTicker(timeout)
	defer ticker.Stop()
	h(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h(ctx)
		}
	}
}

func (t *TimerTask) Start(ctx context.Context, h TimerHandler, timeout time.Duration) {
	t.done = make(DoneChan)
	ctx, t.cancel = context.WithCancel(ctx)
	go t.do(ctx, h, timeout)
}

func (t *TimerTask) Stop() DoneChan {
	t.cancel()
	return t.done
}
