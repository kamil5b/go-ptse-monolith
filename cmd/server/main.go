package main

import (
	"fmt"
	"go-modular-monolith/internal/app/core"
	appHttp "go-modular-monolith/internal/app/http"
	infraMongo "go-modular-monolith/internal/infrastructure/db/mongo"
	infraSQL "go-modular-monolith/internal/infrastructure/db/sql"
)

func main() {
	cfg, err := core.LoadConfig("config/config.yaml")
	if err != nil {
		panic(err)
	}
	featureFlag, err := core.LoadFeatureFlags("config/featureflags.yaml")
	if err != nil {
		panic(err)
	}

	db, err := infraSQL.Open(cfg.App.Database.SQL.DBUrl)
	if err != nil {
		if featureFlag.Repository.Product == "postgres" {
			panic(err)
		}
		fmt.Println("[ERROR] Postgres not loaded:", err)
	}

	mongo, err := infraMongo.OpenMongo(cfg.App.Database.Mongo.MongoURL)
	if err != nil {
		if featureFlag.Repository.Product == "mongo" {
			panic(err)
		}
		fmt.Println("[ERROR] MongoDB not loaded:", err)
	}

	container := core.NewContainer(*featureFlag, db, mongo)
	if container == nil {
		panic("failed to create container")
	}

	switch featureFlag.HTTPHandler {
	case "echo":
		server := appHttp.NewEchoServer(container)
		if err := server.Start(":" + cfg.App.Server.Port); err != nil {
			panic(err)
		}
	default:
		server := appHttp.NewEchoServer(container)
		if err := server.Start(":" + cfg.App.Server.Port); err != nil {
			panic(err)
		}
	}

}
