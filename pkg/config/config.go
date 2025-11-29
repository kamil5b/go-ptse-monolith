package config

import "os"

type Config struct {
	Port      string
	GRPCPort  string
	DBUrl     string
	JWTSecret string
	DBType    string // "postgres" or "mongo"
	MongoURL  string
	MongoDB   string
}

func (c *Config) Load() {
	c = &Config{
		Port:      getenv("PORT", "8080"),
		GRPCPort:  getenv("GRPC_PORT", "9090"),
		DBUrl:     getenv("DATABASE_URL", "postgres://postgres:pass@localhost:5432/app?sslmode=disable"),
		JWTSecret: getenv("JWT_SECRET", "replace-me"),
		DBType:    getenv("DB_TYPE", "postgres"),
		MongoURL:  getenv("MONGO_URL", "mongodb://localhost:27017"),
		MongoDB:   getenv("MONGO_DB", "app"),
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
