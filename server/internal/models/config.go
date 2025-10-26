package models

import (
	"time"

	"github.com/sirupsen/logrus"
)

type (
	ServerConfig struct {
		Port         string
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
	}
	DatabaseConfig struct {
		Host            string
		Port            int
		User            string
		Password        string
		Name            string
		URL             string
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime time.Duration
		ConnMaxIdleTime time.Duration
	}
	AccrualConfig struct {
		Address string
	}
	LoggerConfig struct {
		Level logrus.Level
	}
	JWTConfig struct {
		SecretKey     string
		TokenDuration time.Duration
	}
)
