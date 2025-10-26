package config

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

const (
	EnvProduction  = "production"
	EnvDevelopment = "development"
)

func NewLogger(level logrus.Level, env string) *logrus.Logger {
	logger := logrus.New()

	// Настройка вывода
	var writer io.Writer = os.Stdout
	logger.SetOutput(writer)

	if env == EnvProduction {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
	}
	logger.SetLevel(level)
	return logger
}

func LogConfig(cfg *Config, logger *logrus.Logger) {
	if cfg.Env == EnvDevelopment {
		logger.WithFields(logrus.Fields{
			"env":   cfg.Env,
			"port":  cfg.Server().Port,
			"db":    cfg.Database().Name,
			"host":  cfg.Database().Host,
			"user":  cfg.Database().User,
			"level": cfg.Logger().Level,
		}).Info("Loaded configuration")
	}
}
