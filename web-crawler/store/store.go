package store

import (
	"sync"

	"web-crawler/store/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

type IStore interface {
	sqlc.Querier
}

type Store struct {
	*sqlc.Queries

	logger *logrus.Logger
	mutex  sync.Mutex

	pool *pgxpool.Pool
}

func NewStore(logger *logrus.Logger, pool *pgxpool.Pool) IStore {
	return &Store{
		Queries: sqlc.New(pool),

		logger: logger,

		pool: pool,
	}
}
