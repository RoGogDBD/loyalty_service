package config

import (
	"flag"
	"os"
)

func (c *Config) ParseFlags() {
	var runAddress, databaseURI, accrualAddress string

	flag.StringVar(&runAddress, "a", "", "Address and port to run server")
	flag.StringVar(&databaseURI, "d", "", "Database connection URI")
	flag.StringVar(&accrualAddress, "r", "", "Accrual system address")
	flag.Parse()

	// Флаги > переменные окружения > значения по умолчанию.
	if runAddress != "" {
		c.server.Port = runAddress
	} else if addr := os.Getenv("RUN_ADDRESS"); addr != "" {
		c.server.Port = addr
	}

	if databaseURI != "" {
		c.database.URL = databaseURI
	} else if uri := os.Getenv("DATABASE_URI"); uri != "" {
		c.database.URL = uri
	}

	if accrualAddress != "" {
		c.accrual.Address = accrualAddress
	} else if addr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); addr != "" {
		c.accrual.Address = addr
	}
}
