package config

import (
	"fmt"
	"os"
)

type Config struct {
	MasterDBDSN string
	ServerPort  string
	DBHost      string
	DBUser      string
	DBPassword  string
}

func Load() *Config {
	dbHost := getEnv("DB_HOST", "localhost")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")

	return &Config{
		MasterDBDSN: getEnv("MASTER_DB_DSN",
			fmt.Sprintf("%s:%s@tcp(%s:3306)/master_db?charset=utf8mb4&parseTime=True&loc=Local",
				dbUser, dbPassword, dbHost)),
		ServerPort: getEnv("SERVER_PORT", ":8080"),
		DBHost:     dbHost,
		DBUser:     dbUser,
		DBPassword: dbPassword,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
