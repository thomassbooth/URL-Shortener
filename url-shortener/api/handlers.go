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

// handleJob processes a job by adding it to the worker pool and waiting for the result or timeout.
func (h *Handlers) handleJob(job utils.Job, timeout time.Duration) (string, error) {
	resultChannel := make(chan string)
	job.Result = resultChannel
	h.wp.AddJob(job)

	select {
	case result := <-resultChannel:
		if result == "" {
			return "", http.ErrNoLocation
		}
		return result, nil
	case <-time.After(timeout):
		return "", http.ErrHandlerTimeout
	}
}

// HandleShortenURL handles the POST request to shorten a URL.
func (h *Handlers) HandleShortenURL(w http.ResponseWriter, r *http.Request) {
	// Validate the request method
	if err := utils.ValidateRequest(w, r, http.MethodPost); err != nil {
		return
	}

	// Decode the JSON request payload
	var req struct {
		LongURL string `json:"long_url"`
	}
	if err := utils.DecodeJSON(r.Body, &req); err != nil || req.LongURL == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Process the job
	shortURL, err := h.handleJob(utils.Job{Type: utils.ShortenJob, LongURL: req.LongURL}, 10*time.Second)
	if err == http.ErrNoLocation {
		utils.RespondWithJSON(w, http.StatusNotFound, map[string]string{"message": "No logs found"})
	} else if err == http.ErrHandlerTimeout {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"message": "Timeout while fetching logs"})
	} else {
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"short_url": shortURL})
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

	// Process the job
	longURL, err := h.handleJob(utils.Job{Type: utils.FetchJob, ShortURL: shortURL}, 10*time.Second)
	if err == http.ErrNoLocation {
		utils.RespondWithJSON(w, http.StatusNotFound, map[string]string{"message": "No logs found"})
	} else if err == http.ErrHandlerTimeout {
		utils.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{"message": "Timeout while fetching logs"})
	} else {
		http.Redirect(w, r, longURL, http.StatusFound)
	}
}
