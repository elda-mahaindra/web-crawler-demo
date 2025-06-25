package scheduler

import (
	"fmt"
	"time"

	"web-crawler/service"
	"web-crawler/util/config"

	"github.com/sirupsen/logrus"
)

type Scheduler struct {
	logger *logrus.Logger

	setups []config.SchedulerSetup

	service *service.Service
}

func NewScheduler(
	logger *logrus.Logger,
	setups []config.SchedulerSetup,
	service *service.Service,
) *Scheduler {
	return &Scheduler{
		logger:  logger,
		setups:  setups,
		service: service,
	}
}

// calculateDurationToStartTime calculates how long to wait until the next occurrence of start_time
func (scheduler *Scheduler) calculateDurationToStartTime(startTime string, timezone string) time.Duration {
	const op = "[scheduler] - Scheduler.calculateDurationToStartTime"

	logger := scheduler.logger.WithFields(logrus.Fields{
		"[op]": op,
	})

	logger.Info()

	// Load the specified timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		// Try parsing as UTC offset (e.g., "+07", "-05")
		if offset, offsetErr := time.Parse("-07", timezone); offsetErr == nil {
			// Create a fixed timezone with the parsed offset
			offsetSeconds := int(offset.Sub(time.Time{}).Seconds())

			loc = time.FixedZone(fmt.Sprintf("UTC%s", timezone), offsetSeconds)

			logger.WithFields(logrus.Fields{
				"original_timezone": timezone,
				"parsed_offset":     offsetSeconds / 3600,
				"message":           "Using UTC offset format",
			}).Info()
		} else {
			err := fmt.Errorf("failed to load timezone %s, using UTC: %w", timezone, err)
			logger.WithError(err).Warn()
			loc = time.UTC
		}
	}

	// Get current time in the specified timezone
	now := time.Now().In(loc)

	// Parse the start time (format: "HH:MM") and extract hour/minute
	targetTime, err := time.Parse("15:04", startTime)
	if err != nil {
		err := fmt.Errorf("failed to parse start_time, starting immediately: %w", err)

		logger.WithError(err).Error()

		return 1 * time.Second
	}

	// Create target time for today using the specified timezone
	today := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		targetTime.Hour(),
		targetTime.Minute(),
		0, // seconds
		0, // nanoseconds
		loc,
	)

	logger.WithFields(logrus.Fields{
		"current_time": now.Format("2006-01-02 15:04:05 MST"),
		"target_today": today.Format("2006-01-02 15:04:05 MST"),
		"timezone":     loc.String(),
		"message":      "Calculating start time",
	}).Debug()

	// If the target time has already passed today, schedule for tomorrow
	if today.Before(now) || today.Equal(now) {
		today = today.Add(24 * time.Hour)
		logger.WithFields(logrus.Fields{
			"target_tomorrow": today.Format("2006-01-02 15:04:05 MST"),
			"reason":          "target time has passed today",
			"message":         "Scheduling for tomorrow",
		}).Debug()
	}

	duration := today.Sub(now)

	logger.WithFields(logrus.Fields{
		"final_target":     today.Format("2006-01-02 15:04:05 MST"),
		"duration_seconds": duration.Seconds(),
		"duration_hours":   duration.Hours(),
		"message":          "Calculated duration to start time",
	}).Debug()

	return duration
}
