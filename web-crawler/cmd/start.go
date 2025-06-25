package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"web-crawler/api"
	"web-crawler/scheduler"
	"web-crawler/service"
	"web-crawler/store"
	"web-crawler/util/config"

	"github.com/sirupsen/logrus"
)

func start() {
	const op = "[main] start"

	// --- Init logger ---
	var logger = logrus.New()
	logger.Formatter = new(logrus.JSONFormatter)
	logger.Formatter = new(logrus.TextFormatter)
	logger.Formatter.(*logrus.TextFormatter).DisableColors = true
	logger.Formatter.(*logrus.TextFormatter).DisableTimestamp = true
	logger.Level = logrus.DebugLevel
	logger.Out = os.Stdout

	// --- Load config ---
	config, err := config.LoadConfig(".")
	if err != nil {
		logger.WithFields(logrus.Fields{
			"[op]":  op,
			"scope": "LoadConfig",
			"err":   err.Error(),
		}).Error()

		os.Exit(1)
	}

	logger.WithFields(logrus.Fields{
		"[op]":   op,
		"config": fmt.Sprintf("%+v", config),
	}).Infof("Starting '%s' service ...", config.App.Name)

	// --- Wait for PostgreSQL to be ready ---
	logger.WithFields(logrus.Fields{
		"[op]":    op,
		"message": "Waiting for PostgreSQL to be ready...",
	}).Info()

	time.Sleep(5 * time.Second)

	logger.WithFields(logrus.Fields{
		"[op]":    op,
		"message": "Proceeding to connect to PostgreSQL",
	}).Info()

	// --- Init postgres pool ---
	postgresPool, err := createPostgresPool(logger, config.DB.Postgres)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"[op]":  op,
			"error": err.Error(),
		}).Error()

		os.Exit(1)
	}

	// --- Init store layer ---
	store := store.NewStore(logger, postgresPool)

	// --- Init service layer ---
	service := service.NewService(logger, store)

	// --- Init scheduler ---
	scheduler := scheduler.NewScheduler(logger, config.Scheduler.Setups, service)

	// --- Init api layer ---
	restApi := api.NewApi(logger, service)

	// --- Run scheduler ---
	scheduler.Run()

	// --- Run servers ---
	runRestServer(config.App.Port, restApi)

	// --- Wait for signal ---
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	logger.WithFields(logrus.Fields{
		"[op]":    op,
		"message": "Web crawler service is running. Press Ctrl+C to exit.",
	}).Info()

	// --- Block until signal is received ---
	<-ch

	logger.WithFields(logrus.Fields{
		"[op]":    op,
		"message": "Shutdown signal received, stopping service...",
	}).Info()

	logger.WithFields(logrus.Fields{
		"[op]":    op,
		"message": "Web crawler service stopped gracefully",
	}).Info()
}
