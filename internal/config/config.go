package config

import (
	"log/slog"
	"os"
	"strconv"

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
	config.Crypto.AESKey = getEnv(logger, "CRYPTO_AES_KEY")

	config.SMTP = SMTPConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     getEnvAsInt("SMTP_PORT", 587),
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
		BaseURL:  os.Getenv("APP_BASE_URL"),
	}

	return &config
}

func getEnv(logger *slog.Logger, key string) string {
	value := os.Getenv(key)
	if value == "" {
		logger.Warn("environment variable not set", "env", key)
	}
	return value
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultVal
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultVal
	}
	return value
}
