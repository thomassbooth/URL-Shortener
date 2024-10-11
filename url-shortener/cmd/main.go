package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"url-shortener/url-shortener/api"
	"url-shortener/url-shortener/storage"
)

func main() {
	const connString = "host=localhost port=5432 user=user password=password dbname=url_shortener_db sslmode=disable"
	// Initialize storage (error handling is now internal)
	storage := storage.NewStorage(connString)
	// Notify the channel when an interrupt or terminate signal is received
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	// Create the server
	server := api.NewServer(api.Config{}, storage)
	server.Start()
	// Block the main thread, waiting for a signal to close the application
	sig := <-signals
	log.Printf("Received signal: %v. Shutting down...", sig)
	// Gracefully stop the server and storage
	server.Stop()
	storage.Close()
}
