package internal

import (
	"fmt"
	"url-shortener/url-shortener/storage"
	"url-shortener/url-shortener/utils"
)

type Worker struct {
	id     int
	jobs   <-chan utils.Job
	quit   <-chan struct{}
	active *int32
	store  *storage.Storage
}

type WorkerPool struct {
	jobs    chan utils.Job
	quit    chan struct{}
	workers []*Worker
}

func NewWorkerPool(numWorkers int, store *storage.Storage) *WorkerPool {

	jobs := make(chan utils.Job, 100) // Buffer to hold incoming jobs
	quit := make(chan struct{})       // Channel to signal worker to stop
	pool := &WorkerPool{
		jobs:    jobs,
		quit:    quit,
		workers: make([]*Worker, numWorkers),
	}

	// Setup workers and put them in the pool
	for i := 0; i < numWorkers; i++ {
		worker := Worker{
			id:    i,
			jobs:  jobs,
			quit:  quit,
			store: store,
		}
		pool.workers[i] = &worker
		// Start each worker in a new goroutine
		go worker.start()
	}

	return pool
}
func (w *Worker) start() {
free:
	for {
		select {
		case job := <-w.jobs:
			// Process the job based on its type
			switch job.Type {
			case utils.ShortenJob:
				// Shorten the long URL
				shortURL, err := w.store.ShortenURL(job.LongURL)
				if err != nil {
					fmt.Println("Error shortening URL:", err)
					job.Result <- "" // Send empty string on error
					continue
				}
				job.Result <- shortURL // Send the shortened URL back

			case utils.FetchJob:
				// Fetch the original URL using the short URL
				originalURL, err := w.store.GetOriginalURL(job.ShortURL)
				if err != nil {
					fmt.Println("Error fetching original URL:", err)
					job.Result <- "" // Send empty string on error
					continue
				}
				job.Result <- originalURL // Send the original URL back
			}

		case <-w.quit:
			fmt.Printf("Worker %d stopping\n", w.id)
			w.Stop()
			break free // Exit the loop
		}
	}
}

// Stop stops the worker by sending a signal to its quit channel.
func (w *Worker) Stop() {
}

func (wp *WorkerPool) AddJob(job utils.Job) {
	wp.jobs <- job
}

// Stop stops all workers in the pool.
func (wp *WorkerPool) Stop() {
	// Signal all workers to stop
	for range wp.workers {
		wp.quit <- struct{}{} // Send stop signal to worker
	}

	// Optionally close the jobs channel to prevent further job submissions
	close(wp.jobs) // This is optional
}

// QueuedTasks returns the number of queued tasks.
func (wp *WorkerPool) QueuedTasks() int {
	return len(wp.jobs)
}
