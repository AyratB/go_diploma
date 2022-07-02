package utils

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

func GetConfigs() *Config {
	config := &Config{}

	config.RunAddress = os.Getenv("RUN_ADDRESS")
	if len(config.RunAddress) == 0 {
		flag.StringVar(&config.RunAddress, "a", "localhost:8080", "server run address")
	}
	config.DatabaseURI = os.Getenv("DATABASE_URI")
	if len(config.DatabaseURI) == 0 {
		flag.StringVar(&config.DatabaseURI, "d", "", "db storage path")
	}

	//TODO - какой адрес? как подключать
	config.AccrualSystemAddress = os.Getenv("ACCRUAL_SYSTEM_ADDRESS")
	if len(config.DatabaseURI) == 0 {
		flag.StringVar(&config.DatabaseURI, "r", "", "accrual system address")
	}

	flag.Parse()
	return config
}
