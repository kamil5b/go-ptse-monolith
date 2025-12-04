package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/kamil5b/go-ptse-monolith/internal/app/core"
	appHttp "github.com/kamil5b/go-ptse-monolith/internal/app/http"
	infraMongo "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/db/mongo"
	infraSQL "github.com/kamil5b/go-ptse-monolith/internal/infrastructure/db/sql"
	logger "github.com/kamil5b/go-ptse-monolith/internal/logger"
	productGRPC "github.com/kamil5b/go-ptse-monolith/internal/modules/product/handler/grpc"
	grpctransport "github.com/kamil5b/go-ptse-monolith/internal/transports/grpc"

	"github.com/valyala/fasthttp"
)

func RunServer() error {
	cfg, err := core.LoadConfig("config/config.yaml")
	if err != nil {
		return err
	}
	featureFlag, err := core.LoadFeatureFlags("config/featureflags.yaml")
	if err != nil {
		return err
	}

	db, err := infraSQL.Open(cfg.App.Database.SQL.DBUrl)
	if err != nil {
		if featureFlag.Repository.Product == "postgres" {
			return err
		}
		logger.WithField("error", err).Error("PostgreSQL connection failed")
	}
	defer func() {
		if db != nil {
			db.Close()
		}
	}()

	mongo, err := infraMongo.OpenMongo(cfg.App.Database.Mongo.MongoURL)
	if err != nil {
		if featureFlag.Repository.Product == "mongo" {
			return err
		}
		logger.WithField("error", err).Error("MongoDB connection failed")
	}
	defer func() {
		if mongo != nil {
			infraMongo.CloseMongo(mongo)
		}
	}()

	container := core.NewContainer(*featureFlag, cfg, db, mongo)
	if container == nil {
		return errors.New("failed to create container")
	}

	// Create shutdown context for graceful shutdown
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Start both HTTP and gRPC servers concurrently
	var wg sync.WaitGroup
	errCh := make(chan error, 2)
	var grpcServerInstance *grpctransport.Server

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		switch featureFlag.HTTPHandler {
		case "gin":
			server := appHttp.NewGinServer(container)
			errCh <- server.Run(":" + cfg.App.Server.Port)
		case "nethttp":
			handler := appHttp.NewNetHTTPServer(container)
			errCh <- http.ListenAndServe(":"+cfg.App.Server.Port, handler)
		case "fasthttp":
			handler := appHttp.NewFastHTTPServer(container)
			errCh <- fasthttp.ListenAndServe(":"+cfg.App.Server.Port, handler)
		case "fiber":
			server := appHttp.NewFiberServer(container)
			errCh <- server.Listen(":" + cfg.App.Server.Port)
		default: //default to echo
			server := appHttp.NewEchoServer(container)
			errCh <- server.Start(":" + cfg.App.Server.Port)
		}
	}()

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		grpcServerInstance = grpctransport.NewServer()

		// Register gRPC services
		if container.ProductGRPCHandler != nil {
			grpcServerInstance.RegisterService(productGRPC.RegisterService(container.ProductGRPCHandler))
		}

		logger.WithField("port", cfg.App.Server.GRPCPort).Info("Starting gRPC server")
		if err := grpcServerInstance.Start(shutdownCtx, ":"+cfg.App.Server.GRPCPort); err != nil {
			errCh <- err
		}
	}()

	// Wait for shutdown signal or errors
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Handle signals and errors
	select {
	case <-sigCh:
		logger.Info("Shutdown signal received, starting graceful shutdown...")
		shutdownCancel()

		// Gracefully shutdown worker services
		if container.WorkerServer != nil {
			shutdownWorkerCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := container.WorkerServer.Stop(shutdownWorkerCtx); err != nil {
				logger.WithField("error", err).Error("Error stopping worker server")
			}
			cancel()
		}

		if container.WorkerClient != nil {
			if err := container.WorkerClient.Close(); err != nil {
				logger.WithField("error", err).Error("Error closing worker client")
			}
		}

		// gRPC server stops via context cancellation (shutdownCtx)
		// HTTP server stops via signal handling in framework

		// Wait for servers to finish shutdown with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			logger.Info("Servers stopped successfully")
		case <-time.After(15 * time.Second):
			logger.Warn("Graceful shutdown timeout exceeded, force stopping")
		}

		return nil

	case err := <-errCh:
		if err != nil {
			shutdownCancel()
			return err
		}
	}

	return nil
}
