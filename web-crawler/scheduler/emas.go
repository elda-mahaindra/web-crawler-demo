package scheduler

import (
	"context"
	"fmt"
	"time"

	"web-crawler/service"
	"web-crawler/util/config"

	"github.com/sirupsen/logrus"
)

func (scheduler *Scheduler) RunEmas(setup config.SchedulerSetup) {
	const op = "[scheduler] - Scheduler.RunEmas"

	logger := scheduler.logger.WithFields(logrus.Fields{
		"[op]": op,
	})

	logger.Info()

	ctx := context.TODO()
	ticker := time.NewTicker(setup.TickerDuration)

	for {
		select {
		case <-ctx.Done():
			logger.WithFields(logrus.Fields{
				"message": "context done",
			}).Info()

			return
		case <-ticker.C:
			now := time.Now()

			logger.WithFields(logrus.Fields{
				"now": now,
			}).Info()

			result, err := scheduler.service.CreateEmas(ctx, &service.CreateEmasParams{
				Url:       setup.Url,
				CreatedAt: now,
				Retry: service.RetryConfig{
					MaxAttempts:   setup.Retry.MaxAttempts,
					InitialDelay:  setup.Retry.InitialDelay,
					MaxDelay:      setup.Retry.MaxDelay,
					BackoffFactor: setup.Retry.BackoffFactor,
					EnableJitter:  setup.Retry.EnableJitter,
				},
			})
			if err != nil {
				logger.WithError(err).Error()
			} else {
				logger.WithFields(logrus.Fields{
					"message": "emas created",
					"result":  fmt.Sprintf("%+v", result),
				}).Info()
			}

			ticker.Reset(setup.TickerDuration)
		}
	}
}
