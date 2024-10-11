package api

import (
	"net/http"
	"time"
	"url-shortener/url-shortener/internal"
	"url-shortener/url-shortener/storage"
	"url-shortener/url-shortener/utils"
)

// Handlers struct
type Handlers struct {
	wp *internal.WorkerPool
	db *storage.Storage // Use the new GORM-based Database struct
}

// NewHandlers initializes handlers with the worker pool and GORM database.
func NewHandlers(wp *internal.WorkerPool, db *storage.Storage) *Handlers {
	return &Handlers{wp: wp, db: db}
}

// HandleShortenURL handles the POST request to shorten a URL.
func (h *Handlers) HandleShortenURL(w http.ResponseWriter, r *http.Request) {
	// Validate the request method using the utility function
	if err := utils.ValidateRequest(w, r, http.MethodPost); err != nil {
		return
	}

	var req struct {
		LongURL string `json:"long_url"`
	}

	// Decode the JSON request payload
	if err := utils.DecodeJSON(r.Body, &req); err != nil || req.LongURL == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	// Prepare result channel and add the job to the worker pool
	resultChannel := make(chan string)
	job := utils.Job{Type: utils.ShortenJob, LongURL: req.LongURL, Result: resultChannel}
	h.wp.AddJob(job)

	// Wait for result or timeout
	select {
	case shortURL := <-resultChannel:
		if len(shortURL) == 0 {
			// No result found, respond with a 404
			utils.RespondWithJSON(w, http.StatusNotFound, map[string]string{"message": "No logs found"})
		} else {
			// Successful shortening, return short URL
			utils.RespondWithJSON(w, http.StatusOK, map[string]string{"short_url": shortURL})
		}
	case <-time.After(10 * time.Second):
		// Timeout error handling
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"message": "Timeout while fetching logs"})
	}
}

// HandleRedirect handles the GET request to redirect to the original URL.
func (h *Handlers) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	// Extract the short URL from the path
	shortURL := r.URL.Path[1:]

	if shortURL == "" {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	resultChannel := make(chan string)
	job := utils.Job{Type: utils.FetchJob, ShortURL: shortURL, Result: resultChannel}
	h.wp.AddJob(job)

	// Wait for result or timeout
	select {
	case longURL := <-resultChannel:
		if len(longURL) == 0 {
			// No result found, respond with a 404
			utils.RespondWithJSON(w, http.StatusNotFound, map[string]string{"message": "No logs found"})
		} else {
			// Redirect on success
			http.Redirect(w, r, longURL, http.StatusFound)
		}
	case <-time.After(10 * time.Second):
		// Timeout error handling
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"message": "Timeout while fetching logs"})
	}

}
