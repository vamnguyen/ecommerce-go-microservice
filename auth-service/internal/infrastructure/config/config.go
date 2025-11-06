package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	JWT         JWTConfig
	Security    SecurityConfig
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type SecurityConfig struct {
	MaxLoginAttempts    int
	AccountLockDuration time.Duration
	AllowedOrigins      []string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:            getEnv("PORT", "9001"),
			ReadTimeout:     parseDuration(getEnv("SERVER_READ_TIMEOUT", "10s")),
			WriteTimeout:    parseDuration(getEnv("SERVER_WRITE_TIMEOUT", "10s")),
			ShutdownTimeout: parseDuration(getEnv("SERVER_SHUTDOWN_TIMEOUT", "5s")),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			DBName:          getEnv("DB_NAME", "auth_db"),
			MaxOpenConns:    parseInt(getEnv("DB_MAX_OPEN_CONNS", "25")),
			MaxIdleConns:    parseInt(getEnv("DB_MAX_IDLE_CONNS", "5")),
			ConnMaxLifetime: parseDuration(getEnv("DB_CONN_MAX_LIFETIME", "5m")),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", ""),
			AccessTokenTTL:  parseDuration(getEnv("ACCESS_TOKEN_TTL", "15m")),
			RefreshTokenTTL: parseDuration(getEnv("REFRESH_TOKEN_TTL", "720h")),
		},
		Security: SecurityConfig{
			MaxLoginAttempts:    parseInt(getEnv("MAX_LOGIN_ATTEMPTS", "5")),
			AccountLockDuration: parseDuration(getEnv("ACCOUNT_LOCK_DURATION", "15m")),
			AllowedOrigins:      parseStringSlice(getEnv("ALLOWED_ORIGINS", "http://localhost:3000")),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.JWT.Secret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	return nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DBName,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func parseDuration(s string) time.Duration {
	d, _ := time.ParseDuration(s)
	return d
}

func parseStringSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	return []string{s}
}
