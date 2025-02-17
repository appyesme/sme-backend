package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func Config(key string) string {
	return os.Getenv(key)
}

func SetupEnv() {
	var env_err error

	// When the binary is deployed to VM and run as a unit service, the service will not pick up .env file since the .env file is kept in /apps/{app-name}
	// Therefore we need to load .env from the same location as that of the binary.

	// But in local systems, when we run the app through `air`, it should pick up from the current working directory.
	// Therefore we need to check two locations to load the .env file.
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	env_err = godotenv.Load(filepath.Join(exPath, ".env"))

	if env_err != nil {
		// .env file is not found in the path of executable binary. Check the current directory for .env
		env_err = godotenv.Load(".env")
	}

	if env_err != nil {
		// .env file was not found, check if the code is being run in platforms like Google Cloud Run/DigitalOcean app-platform
		// where env values are directly loaded.
		db_host := Config("DB_HOST")

		if db_host == "" {
			panic("Env variables are missing.")
		}
	}
}
