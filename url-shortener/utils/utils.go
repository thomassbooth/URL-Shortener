package utils

// JobType represents the type of job being processed.
type JobType int

const (
	// ShortenJob indicates a job for shortening a URL.
	ShortenJob JobType = iota
	// FetchJob indicates a job for fetching the original URL.
	FetchJob
)

// Job represents a job that workers will process in the URL shortener.
type Job struct {
	Type     JobType     // Type of the job (shorten or fetch)
	Result   chan string // Channel to send results back (shortened URL or original URL)
	LongURL  string      // Long URL to be shortened (used only for ShortenJob)
	ShortURL string      // Short URL to fetch the original URL (used only for FetchJob)
}
