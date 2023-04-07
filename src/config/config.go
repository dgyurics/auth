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
	Dbname              string
	User                string
	Password            string
	Host                string
	Port                int
	Sslmode             string
	FallbackApplication string // allows for associating database activity with a particular application
	// TODO add support for below items
	// ConnectTimeout       string
	// Sslcert              string
	// Sslkey               string
	// Sslrootcert          string
}

type Redis struct {
	Addr     string
	Username string
	Password string
	DB       int
}

type Config struct {
	ServerConfig ServerConfig
	PostgreSql   PostgreSql
	Redis        Redis
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
			Dbname:              getEnv("POSTGRES_DB", "auth"),
			User:                getEnv("POSTGRES_USER", "postgres"),
			Password:            getEnv("POSTGRES_PASSWORD", "postgres"),
			Host:                getEnv("POSTGRES_HOST", "localhost"),
			Port:                getEnvAsInt("POSTGRES_PORT", 5432),
			Sslmode:             getEnv("POSTGRES_SSLMODE", "disable"),
			FallbackApplication: getEnv("POSTGRES_FALLBACK_APPLICATION", "golang_auth_service"),
		},
		Redis: Redis{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Username: getEnv("REDIS_USERNAME", ""),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
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
