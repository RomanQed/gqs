package gqs

import (
	"math"
	"math/rand/v2"
	"time"
)

type BackoffConfig struct {
	MaxRetries          uint32
	InitialInterval     time.Duration
	MaxInterval         time.Duration
	Multiplier          float64
	RandomizationFactor float64
}

type backoffCounter struct {
	BackoffConfig
}

func (bc *backoffCounter) next(attempt uint32) (time.Duration, bool) {
	if bc.MaxRetries > 0 && attempt > bc.MaxRetries {
		return 0, false
	}
	exp := float64(bc.InitialInterval) * math.Pow(bc.Multiplier, float64(attempt-1))
	if exp > float64(bc.MaxInterval) {
		exp = float64(bc.MaxInterval)
	}
	if bc.RandomizationFactor > 0 {
		delta := bc.RandomizationFactor * exp
		minExp := exp - delta
		maxExp := exp + delta
		exp = minExp + rand.Float64()*(maxExp-minExp)
	}
	return time.Duration(exp), true
}
