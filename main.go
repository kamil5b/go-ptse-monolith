package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kamil5b/go-ptse-monolith/cmd/bootstrap"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		// default: run SQL migrations (up) then start server
		fmt.Println("Running SQL migrations (up) then starting server...")
		if err := bootstrap.RunMigration([]string{"sql", "up"}); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
		if err := bootstrap.RunServer(); err != nil {
			log.Fatalf("server failed: %v", err)
		}
		return
	}

	switch args[0] {
	case "server":
		if err := bootstrap.RunServer(); err != nil {
			log.Fatalf("server failed: %v", err)
		}
	case "worker":
		if err := bootstrap.RunWorker(); err != nil {
			log.Fatalf("worker failed: %v", err)
		}
	case "migration":
		if len(args) < 3 {
			log.Fatalf("usage: go run . migration <sql|mongo> up|down")
		}
		mtype := args[1]
		if mtype != "sql" && mtype != "mongo" {
			log.Fatalf("unsupported migration type: %s (supported: sql,mongo)", mtype)
		}
		action := args[2]
		if action != "up" && action != "down" {
			log.Fatalf("unknown migration action: %s (use up or down)", action)
		}
		if err := bootstrap.RunMigration([]string{mtype, action}); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
	default:
		log.Fatalf("unknown command: %s", args[0])
	}
}
