package env

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	ENV                  string
	DB_CONNECTION_STRING string
	ADMIN_TOKEN          string
)

func GetEnv() error {
	if os.Getenv("ENV") != "staging" && os.Getenv("ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			return err
		}
	}

	ENV = os.Getenv("ENV")
	DB_CONNECTION_STRING = os.Getenv("DB_CONNECTION_STRING")
	ADMIN_TOKEN = os.Getenv("ADMIN_TOKEN")

	return nil
}
