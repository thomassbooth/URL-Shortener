package internal

import (
	"fmt"
	"url-shortener/url-shortener/storage"
	"url-shortener/url-shortener/utils"
)

type Worker struct {
	id    int
	jobs  <-chan utils.Job
	quit  <-chan struct{}
	store *storage.Storage
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
	for {
		select {
		case job := <-w.jobs:
			w.processJob(job) // Delegates job processing to a separate method
		case <-w.quit:
			fmt.Printf("Worker %d stopping\n", w.id)
			return // Exit the loop
		}
	}
}

func (w *Worker) processJob(job utils.Job) {
	var result string
	var err error

	switch job.Type {
	case utils.ShortenJob:
		result, err = w.store.ShortenURL(job.LongURL)
	case utils.FetchJob:
		result, err = w.store.GetOriginalURL(job.ShortURL)
	}

	if err != nil {
		fmt.Println("Error processing job:", err)
		result = "" // Send empty string on error
	}
	job.Result <- result // Send the result back
}

func (wp *WorkerPool) AddJob(job utils.Job) {
	wp.jobs <- job
}

// Stop stops all workers in the pool.
func (wp *WorkerPool) Stop() {
	for range wp.workers {
		wp.quit <- struct{}{} // Send stop signal to worker
	}
	close(wp.jobs) // Close the jobs channel
}

// QueuedTasks returns the number of queued tasks.
func (wp *WorkerPool) QueuedTasks() int {
	return len(wp.jobs)
}
