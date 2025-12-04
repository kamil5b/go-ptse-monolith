package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"go-modular-monolith/internal/app/core"
	logger "go-modular-monolith/internal/logger"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

// RunMigration runs SQL migrations using goose. args is expected to contain the migration action ("up" or "down").
func RunMigrationSQL(args []string) error {
	cfg, err := core.LoadConfig("config/config.yaml")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	migrationsDir := filepath.Join("internal", "infrastructure", "db", "sql", "migration")

	db, err := sql.Open("postgres", cfg.App.Database.SQL.DBUrl)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if len(args) < 1 {
		return fmt.Errorf("missing migration command: up|down")
	}

	switch args[0] {
	case "up":
		if err := goose.Up(db, migrationsDir); err != nil {
			return fmt.Errorf("goose up: %w", err)
		}
	case "down":
		if err := goose.Down(db, migrationsDir); err != nil {
			return fmt.Errorf("goose down: %w", err)
		}
	default:
		return fmt.Errorf("unknown migration command: %s", args[0])
	}

	return nil
}

// RunMigrationMongo executes mongo JS migrations using the `mongosh` CLI.
// It runs all .js files in the mongo migration directory in lexicographical order for "up".
func RunMigrationMongo(args []string) error {
	cfg, err := core.LoadConfig("config/config.yaml")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	migrationsDir := filepath.Join("internal", "infrastructure", "db", "mongo", "migration")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	// collect .js files
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if filepath.Ext(name) == ".js" {
			files = append(files, filepath.Join(migrationsDir, name))
		}
	}
	sort.Strings(files)

	if len(args) < 1 {
		return fmt.Errorf("missing migration command: up|down")
	}
	cmd := args[0]
	switch cmd {
	case "up":
		// run each js with mongosh <uri>/<db> <file>
		mongoURL := cfg.App.Database.Mongo.MongoURL
		dbName := cfg.App.Database.Mongo.MongoDB
		uri := mongoURL
		if dbName != "" {
			// ensure no trailing slash
			uri = fmt.Sprintf("%s/%s", mongoURL, dbName)
		}
		for _, f := range files {
			// run mongosh
			ctx := context.Background()
			// try mongosh first, fallback to mongo if not present
			out, err := exec.CommandContext(ctx, "mongosh", uri, f).CombinedOutput()
			if err != nil {
				// try legacy mongo
				out2, err2 := exec.CommandContext(ctx, "mongo", uri, f).CombinedOutput()
				if err2 != nil {
					return fmt.Errorf("run migration %s failed: %v, mongosh out: %s, mongo out: %s", f, err, string(out), string(out2))
				}
				logger.WithFields(map[string]interface{}{
					"migration": f,
					"output":    string(out2),
				}).Info("MongoDB migration executed (mongo)")
			} else {
				logger.WithFields(map[string]interface{}{
					"migration": f,
					"output":    string(out),
				}).Info("MongoDB migration executed (mongosh)")
			}
		}
	case "down":
		return fmt.Errorf("mongo down migrations are not implemented")
	default:
		return fmt.Errorf("unknown migration command: %s", cmd)
	}
	return nil
}

// RunMigration dispatches migration requests. If args[0] is a type ("sql"|"mongo"), it will use that, otherwise it treats args[0] as the action for SQL.
func RunMigration(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing migration args")
	}
	// if first arg is type
	if args[0] == "sql" || args[0] == "mongo" {
		if len(args) < 2 {
			return fmt.Errorf("missing migration action: up|down")
		}
		mtype := args[0]
		action := args[1]
		switch mtype {
		case "sql":
			return RunMigrationSQL([]string{action})
		case "mongo":
			return RunMigrationMongo([]string{action})
		}
	}
	// otherwise assume SQL and args[0] is action
	return RunMigrationSQL(args)
}
