package storage

import (
	"fmt"
	"log"
	"time"

	"github.com/cenkalti/backoff/v4"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storage struct {
	db *gorm.DB
}

// NewStorage initializes a new GORM connection to PostgreSQL and returns a Storage instance.
func NewStorage(connString string) *Storage {
	db := connectWithRetry(connString)
	migrateDB(db)
	return &Storage{db: db}
}

// connectWithRetry establishes a database connection with exponential backoff.
func connectWithRetry(connString string) *gorm.DB {
	var db *gorm.DB
	var err error

	operation := func() error {
		db, err = gorm.Open(postgres.Open(connString), &gorm.Config{})
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		return nil
	}

	backoffStrategy := backoff.NewExponentialBackOff()
	backoffStrategy.MaxElapsedTime = 1 * time.Minute // Max retry time

	err = backoff.Retry(operation, backoffStrategy)
	if err != nil {
		log.Fatalf("Failed to connect to database after retries: %v", err)
	}

	fmt.Println("Connected to database")
	return db
}

// migrateDB performs auto-migration for the necessary tables.
func migrateDB(db *gorm.DB) {
	if err := db.AutoMigrate(&URL{}); err != nil {
		log.Fatalf("Failed to auto-migrate URL table: %v", err)
	}
}

// Close closes the GORM DB connection (optional with GORM since it uses connection pooling).
func (s *Storage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
