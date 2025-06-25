package service

import (
	"web-crawler/store"

	"github.com/sirupsen/logrus"
)

type Service struct {
	logger *logrus.Logger

	store store.IStore
}

func NewService(
	logger *logrus.Logger,
	store store.IStore,
) *Service {
	return &Service{
		logger: logger,

		store: store,
	}
}
