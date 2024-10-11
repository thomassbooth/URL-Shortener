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

	// Initialize the storage, for some reason I can only do this in the main file
	storage, err := storage.NewStorage(connString)
	if err != nil {
		log.Fatalf("Error connecting to PostgreSQL: %v", err)
	}
	// Notify the channel when an interrupt or terminate signal is received
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	// Create the serverl
	server := api.NewServer(api.Config{}, storage)
	go func() {
		if err := server.Start(); err != nil {
			// this closes our application on Fatal error
			log.Fatalf("Server failed to start: %v", err)
		}
	}()
	// Block the main thread, waiting for a signal to close our application
	sig := <-signals
	log.Printf("Received signal: %v. Shutting down...", sig)
	server.Stop()
	storage.Close()
}
