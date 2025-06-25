package scheduler

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func (service *Scheduler) Run() {
	const op = "[scheduler] Scheduler.Run"

	logger := service.logger.WithFields(logrus.Fields{
		"[op]": op,
	})

	logger.Info()

	for _, setup := range service.setups {
		switch setup.Id {
		case "gold_price":
			{
				go service.RunEmas(setup)
			}
		default:
			err := fmt.Errorf("unrecognized setup id: %s", setup.Id)

			logger.WithError(err).Error()
		}
	}
}
