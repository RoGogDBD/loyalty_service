package config

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"strconv"
	"time"

	"github.com/RoGogDBD/loyalty_service/server/internal/models"
	"github.com/sirupsen/logrus"
)

type (
	Config struct {
		Env      string
		server   models.ServerConfig
		database models.DatabaseConfig
		accrual  models.AccrualConfig
		logger   models.LoggerConfig
		jwt      models.JWTConfig
	}
)

func NewConfig() *Config {
	return &Config{
		server: models.ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			ReadTimeout:  getEnvAsDuration("READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvAsDuration("WRITE_TIMEOUT", 10*time.Second),
		},
		database: models.DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			Name:            getEnv("DB_NAME", "gophermart"),
			URL:             getEnv("DATABASE_URI", ""),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvAsDuration("DB_CONN_MAX_IDLE_TIME", 10*time.Minute),
		},
		accrual: models.AccrualConfig{
			Address: getEnv("ACCRUAL_SYSTEM_ADDRESS", ""),
		},
		logger: models.LoggerConfig{
			Level: getLogLevel(getEnv("LOG_LEVEL", "info")),
		},
		jwt: models.JWTConfig{
			SecretKey:     getEnvOrGenerate("JWT_SECRET"),
			TokenDuration: getEnvAsDuration("JWT_TOKEN_DURATION", 24*time.Hour),
		},
	}
}

func (c *Config) Server() models.ServerConfig     { return c.server }
func (c *Config) Database() models.DatabaseConfig { return c.database }
func (c *Config) Accrual() models.AccrualConfig   { return c.accrual }
func (c *Config) Logger() models.LoggerConfig     { return c.logger }
func (c *Config) JWT() models.JWTConfig           { return c.jwt }

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrGenerate(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	// Генерируем случайный ключ, если не задан
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "default-jwt-secret-change-in-production"
	}
	return base64.URLEncoding.EncodeToString(b)
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return time.Duration(intValue) * time.Second
		}
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getLogLevel(level string) logrus.Level {
	switch level {
	case "debug":
		return logrus.DebugLevel
	case "info":
		return logrus.InfoLevel
	case "warn":
		return logrus.WarnLevel
	case "error":
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}
