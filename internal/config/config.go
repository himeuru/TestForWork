package config

import "os"

type DBConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	DatabaseName string
}

type Config struct {
	Port     string
	DB       DBConfig
	MusicAPI string
}

func LoadConfig() *Config {
	return &Config{
		Port: getEnv("PORT", "8080"),
		DB: DBConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "5432"),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", "admin"),
			DatabaseName: getEnv("DB_NAME", "music_db"),
		},
		MusicAPI: getEnv("MUSIC_API", "http://localhost:8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
