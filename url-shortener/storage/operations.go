package storage

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/big"

	"gorm.io/gorm"
)

// ShortenURL stores the long URL in the database and generates a unique short URL.
func (s *Storage) ShortenURL(longURL string) (string, error) {
	// Check if the long URL already exists in the database
	shortURL, err := s.findExistingShortURL(longURL)
	if err != nil {
		return "", err
	}
	if shortURL != "" {
		return shortURL, nil // Return existing short URL
	}

	// Retry logic to generate a unique short URL
	for retries := 0; retries < 5; retries++ {
		shortURL, err = s.generateUniqueShortURL(longURL)
		if err == nil {
			return shortURL, nil
		}
	}

	return "", fmt.Errorf("failed to generate a unique short URL after multiple attempts")
}

// findExistingShortURL checks if a long URL already exists and returns the corresponding short URL if found.
func (s *Storage) findExistingShortURL(longURL string) (string, error) {
	var existingURL URL
	if err := s.db.Where("long_url = ?", longURL).First(&existingURL).Error; err == nil {
		return existingURL.ShortURL, nil
	}
	return "", nil // Not found
}

// generateUniqueShortURL attempts to generate a unique short URL for a given long URL.
func (s *Storage) generateUniqueShortURL(longURL string) (string, error) {
	nonce, _ := rand.Int(rand.Reader, big.NewInt(1<<32))
	shortURL := generateShortURL(longURL, nonce)
	err := s.shortUrlTransaction(shortURL, longURL)
	return shortURL, err
}

// generateShortURL creates a short URL hash using SHA-1 and returns the first 8 characters.
func generateShortURL(longURL string, nonce *big.Int) string {
	hasher := sha1.New()
	// Concatenate longURL with a nonce to ensure uniqueness
	hasher.Write([]byte(fmt.Sprintf("%s%d", longURL, nonce)))
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	return hashString[:8] // Return the first 8 characters of the hash
}

// shortUrlTransaction inserts the short URL and long URL into the database.
func (s *Storage) shortUrlTransaction(shortURL, longURL string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var existingURL URL
		if err := tx.Where("short_url = ?", shortURL).First(&existingURL).Error; err == nil {
			// Short URL already exists
			if existingURL.LongURL == longURL {
				return nil // The longURL matches, so no need to insert
			}
			return fmt.Errorf("short URL exists but with a different long URL")
		} else if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to query for existing short URL: %v", err)
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
}

// GetOriginalURL retrieves the long URL from the database given a short URL.
func (s *Storage) GetOriginalURL(shortURL string) (string, error) {
	var url URL
	if err := s.db.First(&url, "short_url = ?", shortURL).Error; err != nil {
		return "", fmt.Errorf("failed to find original URL: %v", err)
	}

	return url.LongURL, nil
}
