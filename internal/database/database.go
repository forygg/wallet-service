package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"os"

	"wallet-service/internal/config"
	_ "github.com/lib/pq"
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.Database.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Настраиваем connection pool
	db.SetMaxOpenConns(25)                 // Максимум соединений
	db.SetMaxIdleConns(10)                 // Максимум idle соединений  
	db.SetConnMaxLifetime(5 * time.Minute) // Максимальное время жизни соединения
	db.SetConnMaxIdleTime(2 * time.Minute) // Максимальное время idle

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to database with connection pool")
	return db, nil
}

func RunMigrations(db *sql.DB) error {
	// Читаем миграцию из файла
	migrationSQL, err := os.ReadFile("migrations/001_create_wallets.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	_, err = db.Exec(string(migrationSQL))
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	
	log.Println("Migrations completed successfully")
	return nil
}