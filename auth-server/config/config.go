package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// ServerConfig contains configuration values for the server listening port.
type ServerConfig struct {
	Port string
}

// PostgreSQL contains configuration values for the PostgreSQL database.
type PostgreSQL struct {
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

// Cors contains configuration values for the CORS middleware.
type Cors struct {
	AllowOrigin      string
	AllowMethods     string
	AllowHeaders     string
	AllowCredentials string
}

// Redis contains configuration values for the Redis cache.
type Redis struct {
	Addr     string
	Username string
	Password string
	DB       int
}

// Session contains configuration values for the session.
type Session struct {
	Name     string
	Domain   string
	Path     string
	Secure   bool
	HTTPOnly bool
	SameSite string
	MaxAge   int
}

// RequestTimeout contains configuration value for http request timeout.
type RequestTimeout int

// Config is the container struct for all configuration values.
type Config struct {
	Cors
	PostgreSQL
	Redis
	RequestTimeout
	ServerConfig
	Session
}

func init() {
	_ = godotenv.Load()
}

// New returns a configuration struct with default values
// and environment variables overriding the defaults.
func New() Config {
	return Config{
		Cors: Cors{
			AllowOrigin:      getEnv("CORS_ALLOW_ORIGIN", "*"),
			AllowMethods:     getEnv("CORS_ALLOW_METHODS", "GET, POST, OPTIONS"),
			AllowHeaders:     getEnv("CORS_ALLOW_HEADERS", "*"),
			AllowCredentials: getEnv("CORS_ALLOW_CREDENTIALS", "true"),
		},
		PostgreSQL: PostgreSQL{
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
		RequestTimeout: RequestTimeout(getEnvAsInt("REQUEST_TIMEOUT", 30)),
		ServerConfig: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Session: Session{
			Name:     getEnv("SESSION_NAME", "X-Session-ID"),
			Domain:   getEnv("SESSION_DOMAIN", "localhost"),
			Path:     getEnv("SESSION_PATH", "/"),
			Secure:   getEnvAsBool("SESSION_SECURE", false),
			HTTPOnly: getEnvAsBool("SESSION_HTTP_ONLY", true),
			SameSite: getEnv("SESSION_SAME_SITE", "Strict"),
			MaxAge:   getEnvAsInt("SESSION_MAX_AGE", 86400),
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
