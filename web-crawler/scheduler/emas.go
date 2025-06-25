package scheduler

import (
	"context"
	"time"

	"web-crawler/service"
	"web-crawler/util/config"

	"github.com/sirupsen/logrus"
)

func (scheduler *Scheduler) RunEmas(setup config.SchedulerSetup) {
	const op = "[scheduler] - Scheduler.RunEmas"
	const maxConcurrentJobs = 3

	logger := scheduler.logger.WithFields(logrus.Fields{
		"[op]":     op,
		"setup_id": setup.Id,
	})

	logger.WithFields(logrus.Fields{
		"message":    "Starting scheduler",
		"start_time": setup.StartTime,
		"duration":   setup.TickerDuration,
	}).Info()

	ctx := context.TODO()

	// Semaphore to limit concurrent scraping jobs
	jobSemaphore := make(chan struct{}, maxConcurrentJobs)

	// Calculate initial delay to reach start_time
	initialDelay := scheduler.calculateDurationToStartTime(setup.StartTime, setup.Timezone)

	logger.WithFields(logrus.Fields{
		"initial_delay_seconds": initialDelay.Seconds(),
	}).Info()

	// Create ticker with initial delay
	ticker := time.NewTicker(initialDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.WithFields(logrus.Fields{
				"message": "context done, stopping scheduler",
			}).Info()

			return

		case tickTime := <-ticker.C:
			logger.WithFields(logrus.Fields{
				"tick_time": tickTime.Format("2006-01-02 15:04:05"),
			}).Info()

			// Reset ticker immediately with new duration
			ticker.Reset(setup.TickerDuration)

			logger.WithFields(logrus.Fields{
				"next_tick_in_seconds": setup.TickerDuration.Seconds(),
				"next_tick_time":       time.Now().Add(setup.TickerDuration).Format("2006-01-02 15:04:05"),
			}).Info()

			// Execute scraping logic in goroutine (non-blocking)
			go func(tickTime time.Time) {
				const op = "[scheduler] - Scheduler.RunEmas - scraping job"

				logger := scheduler.logger.WithFields(logrus.Fields{
					"[op]": op,
				})

				// Acquire semaphore (blocks if max concurrent jobs reached)
				select {
				case jobSemaphore <- struct{}{}:
					// Got semaphore, proceed
				case <-ctx.Done():
					logger.WithFields(logrus.Fields{
						"message": "context cancelled while waiting for job slot",
					}).Warn()

					return
				}

				// Release semaphore when done
				defer func() { <-jobSemaphore }()

				logger.WithFields(logrus.Fields{
					"job_start_time": tickTime.Format("2006-01-02 15:04:05"),
				}).Info()

				jobStartTime := time.Now()

				// Execute the scraping
				result, err := scheduler.service.CreateEmas(ctx, &service.CreateEmasParams{
					Url:       setup.Url,
					CreatedAt: tickTime,
					Retry: service.RetryConfig{
						MaxAttempts:   setup.Retry.MaxAttempts,
						InitialDelay:  setup.Retry.InitialDelay,
						MaxDelay:      setup.Retry.MaxDelay,
						BackoffFactor: setup.Retry.BackoffFactor,
						EnableJitter:  setup.Retry.EnableJitter,
					},
				})

				jobDuration := time.Since(jobStartTime)

				if err != nil {
					logger.WithFields(logrus.Fields{
						"error":                err.Error(),
						"job_duration_seconds": jobDuration.Seconds(),
					}).Error()
				} else {
					logger.WithFields(logrus.Fields{
						"emas_id":              result.ID,
						"job_duration_seconds": jobDuration.Seconds(),
					}).Info()
				}
			}(tickTime)
		}
	}
}
