package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Port string
}

type PostgreSql struct {
	Url string
	// Host     string
	// Port     int
	// User     string
	// Password string
	// DbName   string
}

type Config struct {
	ServerConfig ServerConfig
	PostgreSql   PostgreSql
}

func init() {
	_ = godotenv.Load()
}

func New() *Config {
	return &Config{
		ServerConfig: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		PostgreSql: PostgreSql{
			Url: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/auth"),
		},
	}
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := getEnv(name, "")
	if valStr == "" {
		return defaultVal
	}
	val := strings.Split(valStr, sep)
	return val
}
