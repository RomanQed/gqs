package gqs

import (
	"errors"
	"github.com/romanqed/gqs/internal"
	"sync/atomic"
	"time"
)

const (
	stopped = iota
	started
)

var (
	// ErrDoubleStarted is returned when Start is called on a worker that
	// has already been started.
	//
	// Workers managed by gqs follow a strict lifecycle and must not be
	// started more than once without being stopped.
	ErrDoubleStarted = errors.New("worker double start")

	// ErrDoubleStopped is returned when Stop is called on a worker that
	// is not currently running.
	ErrDoubleStopped = errors.New("worker double stop")

	// ErrStopTimeout is returned when a worker fails to shut down within
	// the provided timeout during Stop.
	//
	// In this case, the worker may still be terminating in the background.
	ErrStopTimeout = errors.New("worker stop timeout")
)

type lcBase struct {
	state atomic.Int32
}

func (lb *lcBase) tryStart() error {
	if !lb.state.CompareAndSwap(stopped, started) {
		return ErrDoubleStarted
	}
	return nil
}

func (lb *lcBase) tryStop(timeout time.Duration, df internal.DoneFunc) error {
	if !lb.state.CompareAndSwap(started, stopped) {
		return ErrDoubleStopped
	}
	done := df()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
		return nil
	case <-timer.C:
		return ErrStopTimeout
	}
}
