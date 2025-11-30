package bootstrap

import (
	"errors"
	"fmt"
	"go-modular-monolith/internal/app/core"
	appHttp "go-modular-monolith/internal/app/http"
	infraMongo "go-modular-monolith/internal/infrastructure/db/mongo"
	infraSQL "go-modular-monolith/internal/infrastructure/db/sql"
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
		fmt.Println("[ERROR] Postgres not loaded:", err)
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
		fmt.Println("[ERROR] MongoDB not loaded:", err)
	}
	defer func() {
		if mongo != nil {
			infraMongo.CloseMongo(mongo)
		}
	}()

	container := core.NewContainer(*featureFlag, db, mongo)
	if container == nil {
		return errors.New("failed to create container")
	}

	switch featureFlag.HTTPHandler {
	case "gin":
		server := appHttp.NewGinServer(container)
		if err := server.Run(":" + cfg.App.Server.Port); err != nil {
			return err
		}
	default: //default to echo
		server := appHttp.NewEchoServer(container)
		if err := server.Start(":" + cfg.App.Server.Port); err != nil {
			return err
		}
	}
	return nil

}
