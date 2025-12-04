package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kamil5b/go-ptse-monolith/internal/app/core"
	"github.com/kamil5b/go-ptse-monolith/internal/app/worker"
	infraMongo "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/db/mongo"
	infraSQL "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/db/sql"
	logger "github.com/kamil5b/go-ptse-monolith/internal/logger"
	userworker "github.com/kamil5b/go-ptse-monolith/internal/modules/user/worker"
)

// RunWorker initializes and starts the worker server
func RunWorker() error {
	cfg, err := core.LoadConfig("config/config.yaml")
	if err != nil {
		return err
	}

	featureFlag, err := core.LoadFeatureFlags("config/featureflags.yaml")
	if err != nil {
		return err
	}

	// Check if workers are enabled
	if !featureFlag.Worker.Enabled || featureFlag.Worker.Backend == "disable" {
		return errors.New("workers are not enabled")
	}

	// Initialize databases
	db, err := infraSQL.Open(cfg.App.Database.SQL.DBUrl)
	if err != nil {
		if featureFlag.Repository.User == "postgres" || featureFlag.Repository.Product == "postgres" || featureFlag.Repository.Authentication == "postgres" {
			return err
		}
		logger.WithField("error", err).Warn("PostgreSQL connection failed")
	}
	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	mongo, err := infraMongo.OpenMongo(cfg.App.Database.Mongo.MongoURL)
	if err != nil {
		if featureFlag.Repository.Product == "mongo" || featureFlag.Repository.Authentication == "mongo" {
			return err
		}
		logger.WithField("error", err).Warn("MongoDB connection failed")
	}
	defer func() {
		if mongo != nil {
			infraMongo.CloseMongo(mongo)
		}
	}()

	// Create container with all dependencies
	container := core.NewContainer(*featureFlag, cfg, db, mongo)
	if container == nil {
		return errors.New("failed to create container")
	}

	// Initialize worker manager
	workerManager := worker.NewWorkerManager(container)

	// Setup module task registrations
	logger.Info("Setting up task registrations...")
	moduleRegistry := worker.NewModuleRegistry()

	// Register user module tasks (module only provides definitions, no app imports)
	moduleRegistry.Register(userworker.NewUserModuleWorkerTasks())

	// Register all module tasks with the task registry
	if err := moduleRegistry.RegisterAllTasks(
		workerManager.GetRegistry(),
		container.UserRepository,
		container.EmailClient,
		featureFlag.Worker.Tasks.EmailNotifications,
		featureFlag.Worker.Tasks.DataExport,
		featureFlag.Worker.Tasks.ReportGeneration,
	); err != nil {
		return fmt.Errorf("failed to register module tasks: %w", err)
	}

	// Register all collected tasks with the worker server
	logger.Info("Registering all tasks with worker server...")
	if err := workerManager.RegisterTasks(); err != nil {
		return fmt.Errorf("failed to register tasks: %w", err)
	}

	// Register all cron jobs from modules
	logger.Info("Registering cron jobs...")
	if err := moduleRegistry.RegisterAllCronJobs(
		workerManager.GetCronScheduler(),
		featureFlag.Worker.Tasks.EmailNotifications,
	); err != nil {
		return fmt.Errorf("failed to register cron jobs: %w", err)
	}

	// Create a context that can be canceled by signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start the worker server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		logger.WithField("backend", featureFlag.Worker.Backend).Info("Worker server running")
		errChan <- workerManager.Start(ctx)
	}()

	// Wait for either a signal or an error
	select {
	case sig := <-sigChan:
		logger.WithField("signal", sig).Info("Received signal, initiating graceful shutdown")
		cancel()
		// Give the server a moment to shut down cleanly
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		if err := workerManager.Stop(shutdownCtx); err != nil {
			logger.WithField("error", err).Warn("Error during worker shutdown")
		}
		logger.Info("Worker server stopped")
		return nil
	case err := <-errChan:
		logger.WithField("error", err).Error("Worker server error")
		return err
	}
}
