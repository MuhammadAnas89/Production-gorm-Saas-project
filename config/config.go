package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerPort string

	// Master Database Connection String
	MasterDBDSN string

	// Tenant Database Details (Taaki hum code mein dynamic connection bana saken)
	DBHost     string
	DBUser     string
	DBPassword string

	// Redis
	RedisAddr string
	RedisPass string
}

func Load() *Config {
	// Defaults set kar rahe hain agar env file na ho
	dbHost := getEnv("DB_HOST", "localhost")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")

	// Master DB ka DSN default banate hain
	defaultMasterDSN := fmt.Sprintf("%s:%s@tcp(%s:3306)/master_db?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost)

	return &Config{
		ServerPort:  getEnv("SERVER_PORT", ":8080"),
		MasterDBDSN: getEnv("MASTER_DB_DSN", defaultMasterDSN),

		DBHost:     dbHost,
		DBUser:     dbUser,
		DBPassword: dbPassword,

		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass: getEnv("REDIS_PASSWORD", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
