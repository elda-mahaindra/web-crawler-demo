package config

import (
	"time"
)

// App config

type App struct {
	Name string `mapstructure:"name"`
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DB config

type PostgresPool struct {
	MaxConns int `mapstructure:"max_conns"`
	MinConns int `mapstructure:"min_conns"`
}

type PostgresConfig struct {
	ConnectionString string       `mapstructure:"connection_string"`
	Pool             PostgresPool `mapstructure:"pool"`
}

type DB struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
}

// Scheduler config

type SchedulerSetup struct {
	Id             string        `mapstructure:"id"`
	Url            string        `mapstructure:"url"`
	StartTime      string        `mapstructure:"start_time"`
	TickerDuration time.Duration `mapstructure:"ticker_duration"`
	Timezone       string        `mapstructure:"timezone"`
	Retry          RetryConfig   `mapstructure:"retry"`
}

type RetryConfig struct {
	MaxAttempts   int           `mapstructure:"max_attempts"`
	InitialDelay  time.Duration `mapstructure:"initial_delay"`
	MaxDelay      time.Duration `mapstructure:"max_delay"`
	BackoffFactor float64       `mapstructure:"backoff_factor"`
	EnableJitter  bool          `mapstructure:"enable_jitter"`
}

type Scheduler struct {
	Setups []SchedulerSetup `mapstructure:"setups"`
}
