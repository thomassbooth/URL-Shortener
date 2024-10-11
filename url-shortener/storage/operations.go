package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"gorm.io/gorm"
)

// ShortenURL stores the long URL in the database and generates a unique short URL.
func (s *Storage) ShortenURL(longURL string) (string, error) {
	var shortURL string
	var err error

	// Retry logic
	for retries := 0; retries < 5; retries++ {
		// Generate a short URL hash
		shortURL = generateShortURL(longURL)

		var existingURL URL
		if err := s.db.Where("long_url = ?", longURL).First(&existingURL).Error; err == nil {
			// If it exists, return the existing short URL
			return existingURL.ShortURL, nil
		}
		// Insert the URL entry into the database in a transaction
		err = s.shortUrlTransaction(shortURL, longURL)
		// Check if the transaction was successful
		if err == nil {
			return shortURL, nil
		}

		// If the error is "short URL already exists", regenerate and try again
		if err.Error() != "short URL already exists" {
			return "", err
		}
	}

	return "", fmt.Errorf("failed to generate a unique short URL after multiple attempts")
}

// GetOriginalURL retrieves the long URL from the database given a short URL.
func (s *Storage) GetOriginalURL(shortURL string) (string, error) {
	var url URL

	// Retrieve the long URL by its short URL
	if err := s.db.First(&url, "short_url = ?", shortURL).Error; err != nil {
		return "", fmt.Errorf("failed to find original URL: %v", err)
	}

	return url.LongURL, nil
}

// generateShortURL creates a short URL hash using SHA-1 and returns the first 6 characters.
func generateShortURL(longURL string) string {
	hasher := sha1.New()
	hasher.Write([]byte(longURL))
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	return hashString[:8]
}

func (s *Storage) shortUrlTransaction(shortURL, longURL string) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Check if the short URL already exists
		var existingURL URL
		if err := tx.Where("short_url = ?", shortURL).First(&existingURL).Error; err == nil {
			// Short URL already exists, return a conflict error
			return fmt.Errorf("short URL already exists")
		}

		// Insert the new URL entry
		url := URL{
			ShortURL: shortURL,
			LongURL:  longURL,
		}
		if err := tx.Create(&url).Error; err != nil {
			return fmt.Errorf("failed to insert URL into database: %v", err)
		}

		return nil // Successful transaction
	})

	return err
}
