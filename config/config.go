package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerPort  string
	MasterDBDSN string

	// Database Connection Params
	DBHost     string
	DBUser     string
	DBPassword string

	RedisAddr string
	RedisPass string

	// ✅ JWT Secret added here
	JWTSecret string
}

func Load() *Config {
	dbHost := getEnv("DB_HOST", "localhost")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")

	defaultMasterDSN := fmt.Sprintf("%s:%s@tcp(%s:3306)/master_db?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost)

	return &Config{
		ServerPort:  getEnv("SERVER_PORT", ":8080"),
		MasterDBDSN: getEnv("MASTER_DB_DSN", defaultMasterDSN),
		DBHost:      dbHost,
		DBUser:      dbUser,
		DBPassword:  dbPassword,
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:   getEnv("REDIS_PASSWORD", ""),
		// ✅ Default secret for dev, change in prod
		JWTSecret: getEnv("JWT_SECRET", "super_secret_key_change_me_in_prod"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
