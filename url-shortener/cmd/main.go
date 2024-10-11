package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"url-shortener/url-shortener/api"
	"url-shortener/url-shortener/storage"
)

var config = api.Config{
	ListenAddr: ":8005",
	DSN:        "host=db port=5432 user=user password=password dbname=url_shortener_db sslmode=disable",
	Rate:       10,
	Burst:      20,
}

func main() {
	// Initialize storage (error handling is now internal)
	storage := storage.NewStorage(config.DSN)
	// Notify the channel when an interrupt or terminate signal is received
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	// Create the server
	server := api.NewServer(config, storage)
	server.Start()
	// Block the main thread, waiting for a signal to close the application
	sig := <-signals
	log.Printf("Received signal: %v. Shutting down...", sig)
	// Gracefully stop the server and storage
	server.Stop()
	storage.Close()
}
