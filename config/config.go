package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	DB_NAME string
	DB_USER string
	DB_PASS string
	DB_HOST string
	DB_PORT string

	ACCESS_TOKEN    string
	REFRESH_TOKEN   string
	EXPIRED_ACCESS  string
	EXPIRED_REFRESH string

	RUN_PORT string
}

func NewConfig() Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	config := Config{}

	config.DB_NAME = os.Getenv("DB_NAME")
	config.DB_USER = os.Getenv("DB_USER")
	config.DB_PASS = os.Getenv("DB_PASS")
	config.DB_HOST = os.Getenv("DB_HOST")
	config.DB_PORT = os.Getenv("DB_PORT")

	config.RUN_PORT = os.Getenv("RUN_PORT")

	config.ACCESS_TOKEN = os.Getenv("ACCESS_TOKEN")
	config.REFRESH_TOKEN = os.Getenv("REFRESH_TOKEN")
	config.EXPIRED_ACCESS = os.Getenv("EXPIRED_ACCESS")
	config.EXPIRED_REFRESH = os.Getenv("EXPIRED_REFRESH")

	return config
}
