package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Log      LogConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Path string
}

type JWTConfig struct {
	Secret     string
	ExpireHour int
}

type LogConfig struct {
	Level      string
	OutputPath string
	MaxSize    int // megabytes
	MaxBackups int
	MaxAge     int // days
}

type CORSConfig struct {
	AllowOrigins []string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:  time.Second * time.Duration(getEnvInt("SERVER_READ_TIMEOUT", 30)),
			WriteTimeout: time.Second * time.Duration(getEnvInt("SERVER_WRITE_TIMEOUT", 30)),
		},
		Database: DatabaseConfig{
			Path: getEnv("DB_PATH", "./data/shiny.db"),
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "shiny-collection-secret-key-change-in-production"),
			ExpireHour: getEnvInt("JWT_EXPIRE_HOUR", 720),
		},
		Log: LogConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			OutputPath: getEnv("LOG_OUTPUT", "./logs/app.log"),
			MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
			MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 3),
			MaxAge:     getEnvInt("LOG_MAX_AGE", 28),
		},
		CORS: CORSConfig{
			AllowOrigins: []string{getEnv("CORS_ORIGIN", "http://localhost:5173")},
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
