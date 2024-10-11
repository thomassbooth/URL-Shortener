package storage

import "time"

type URL struct {
	ShortURL  string     `gorm:"primaryKey;unique;not null"` // Ensures ShortURL is unique and not null
	LongURL   string     `gorm:"not null;index"`             // Long URL must be present and indexed
	CreatedAt time.Time  `gorm:"autoCreateTime"`             // Automatically set on creation
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`             // Automatically updates on each save
	ExpiresAt *time.Time `gorm:"default:null"`               // Optional expiration timestamp
}
