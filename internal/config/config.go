package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port		string
	Database	DatabaseConfig
}

type DatabaseConfig struct {
	Host		string
	Port		string
	User		string
	Password	string
	Name		string
	SSLMode		string
}

func Load() (*Config, error) {
    cfg := &Config{
        Port: getEnv("PORT", "8080"),
        Database: DatabaseConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     getEnv("DB_PORT", "5432"),
            User:     getEnv("DB_USER", "wallet_user"),
            Password: getEnv("DB_PASSWORD", "wallet_password"),
            Name:     getEnv("DB_NAME", "wallet_db"),
            SSLMode:  getEnv("DB_SSLMODE", "disable"),
        },
    }
    
    return cfg, nil
}

func (d DatabaseConfig) ConnectionString() string {
    return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}