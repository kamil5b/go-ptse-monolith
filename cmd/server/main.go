package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/mongo"

	"go-modular-monolith/internal/auth"
	"go-modular-monolith/internal/product"
	"go-modular-monolith/pkg/config"
	"go-modular-monolith/pkg/db"
	"go-modular-monolith/pkg/logger"
)

func main() {
	cfg := config.Config{}
	cfg.Load()
	logger.Init()

	var (
		sqlDB       *sqlx.DB
		mongoClient *mongo.Client
	)

	// choose DB based on config
	dbType := cfg.DBType
	var authService *auth.Service
	var productRepo product.ProductRepository

	switch strings.ToLower(dbType) {
	case "mongo":
		// open mongo for products
		mc, err := db.OpenMongo(cfg.MongoURL)
		if err != nil {
			logger.Log.Fatal().Err(err).Msg("failed to open mongo")
		}
		mongoClient = mc
		productRepo = product.NewMongoRepository(mc, cfg.MongoDB)
		// try to open Postgres for auth if DATABASE_URL provided
		if cfg.DBUrl != "" {
			d, err := db.Open(cfg.DBUrl)
			if err == nil {
				sqlDB = d
				authService = auth.NewService(cfg.JWTSecret, d)
			} else {
				logger.Log.Warn().Err(err).Msg("failed to open postgres for auth; auth routes disabled")
			}
		}
	default:
		// default to postgres
		d, err := db.Open(cfg.DBUrl)
		if err != nil {
			logger.Log.Fatal().Err(err).Msg("failed to open db")
		}
		sqlDB = d
		authService = auth.NewService(cfg.JWTSecret, d)
		productRepo = product.NewSQLRepository(d)
	}

	productSvc := product.NewService(productRepo)

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	// auth routes (only if authService initialized)
	if authService != nil {
		a := e.Group("/v1/auth")
		a.POST("/register", auth.NewHandler(authService).Register)
		a.GET("/activate", auth.NewHandler(authService).Activate)
		a.POST("/login", auth.NewHandler(authService).Login)
		a.POST("/forgot", auth.NewHandler(authService).ForgotPassword)
		a.POST("/reset", auth.NewHandler(authService).ResetPassword)
	} else {
		logger.Log.Warn().Msg("auth service not initialized; auth routes are disabled")
	}

	// product routes
	p := e.Group("/v1/products")
	if authService != nil {
		p.Use(auth.JWTMiddleware(authService))
	} else {
		logger.Log.Warn().Msg("product routes are unprotected (auth disabled)")
	}
	ph := product.NewHandler(productSvc)
	p.POST("/", ph.Create)
	p.GET("/", ph.List)
	p.GET("/:id", ph.Get)
	p.PUT("/:id", ph.Update)
	p.DELETE("/:id", ph.Delete)

	addr := ":" + cfg.Port
	// start server
	go func() {
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal().Err(err).Msg("http server failed")
		}
	}()

	fmt.Println("Server started on", addr)

	// wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	logger.Log.Info().Str("signal", sig.String()).Msg("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.Log.Error().Err(err).Msg("error during http server shutdown")
	}

	if sqlDB != nil {
		if err := sqlDB.Close(); err != nil {
			logger.Log.Error().Err(err).Msg("error closing sql db")
		}
	}
	if mongoClient != nil {
		if err := db.CloseMongo(mongoClient); err != nil {
			logger.Log.Error().Err(err).Msg("error closing mongo client")
		}
	}

	logger.Log.Info().Msg("server stopped")
}
