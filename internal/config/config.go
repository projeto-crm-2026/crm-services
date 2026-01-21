package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

func LoadConfigs(logger *slog.Logger) *Config {
	godotenv.Load(".env")

	var config Config

	config.Server.Address = getEnv(logger, "SERVER_ADDRESS")

	config.DB.User = getEnv(logger, "DB_USER")
	config.DB.Password = getEnv(logger, "DB_PASSWORD")
	config.DB.Address = getEnv(logger, "DB_ADDRESS")
	config.DB.Port = getEnv(logger, "DB_PORT")
	config.DB.Name = getEnv(logger, "DB_NAME")
	config.JWT.JWTSecret = getEnv(logger, "JWT_SECRET")

	return &config
}

func getEnv(logger *slog.Logger, key string) string {
	value := os.Getenv(key)
	if value == "" {
		logger.Warn("environment variable not set", "env", key)
	}
	return value
}
