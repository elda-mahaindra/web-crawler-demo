package scheduler

import (
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
		logger: logger,

		setups: setups,

		service: service,
	}
}
