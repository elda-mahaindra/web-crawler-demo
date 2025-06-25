package service

import (
	"math"
	"math/rand"
	"time"
)

type RetryConfig struct {
	MaxAttempts   int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
	EnableJitter  bool
}

// calculateBackoffDelay calculates the delay for exponential backoff with optional jitter
func (service *Service) calculateBackoffDelay(attempt int, retryConfig RetryConfig) time.Duration {
	// Calculate exponential backoff delay
	exponentialDelay := time.Duration(float64(retryConfig.InitialDelay) * math.Pow(retryConfig.BackoffFactor, float64(attempt-1)))

	// Cap the delay at MaxDelay
	if exponentialDelay > retryConfig.MaxDelay {
		exponentialDelay = retryConfig.MaxDelay
	}

	// Add jitter if enabled (±25% randomization)
	if retryConfig.EnableJitter {
		jitterRange := float64(exponentialDelay) * 0.25
		jitter := time.Duration(rand.Float64()*jitterRange*2 - jitterRange)
		exponentialDelay += jitter

		// Ensure delay is not negative
		if exponentialDelay < 0 {
			exponentialDelay = retryConfig.InitialDelay
		}
	}

	return exponentialDelay
}
