package storage

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// URL represents the URL entity in the database.
type URL struct {
	ShortURL  string `gorm:"primaryKey"`
	LongURL   string
	CreatedAt time.Time
}

// Storage struct holds the GORM DB instance.
type Storage struct {
	db *gorm.DB
}

// NewStorage initializes a new GORM connection to PostgreSQL and returns a Storage instance.
func NewStorage(connString string) (*Storage, error) {
	fmt.Println("connString: ", connString)
	db, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Auto-migrate the URL struct to create or update the table schema
	err = db.AutoMigrate(&URL{})
	if err != nil {
		return nil, fmt.Errorf("failed to auto-migrate URL table: %v", err)
	}

	return &Storage{db: db}, nil
}

// Close closes the GORM DB connection (optional with GORM since it uses connection pooling).
func (s *Storage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
