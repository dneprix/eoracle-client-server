package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"eoracle-client-server/internal/output"
	"eoracle-client-server/internal/queue"
	"eoracle-client-server/internal/server"
)

func main() {
	var (
		rabbitURL      = flag.String("rabbit-url", "amqp://guest:guest@localhost:5672/", "RabbitMQ URL")
		readQueueName  = flag.String("read-queue-name", "read-commands", "Queue name for read commands")
		writeQueueName = flag.String("write-queue-name", "write-commands", "Queue name for write commands")

		outputFileName = flag.String("output-file", "server_output.txt", "Output file for results")
		readWorkers    = flag.Int("read-workers", 100, "Number of read worker goroutines")
		writeWorkers   = flag.Int("write-workers", 10, "Number of write worker goroutines")
	)
	flag.Parse()

	// Initialize context with cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Create queue for read commands
	readQueue, err := queue.NewRabbitMQQueue(*rabbitURL, *readQueueName)
	if err != nil {
		log.Fatalf("Failed to create queue for read commands: %v", err)
	}
	defer readQueue.Close()

	// Create queue for write commands
	writeQueue, err := queue.NewRabbitMQQueue(*rabbitURL, *writeQueueName)
	if err != nil {
		log.Fatalf("Failed to create queue for write commands: %v", err)
	}
	defer writeQueue.Close()

	// Create output file for results
	outputFile, err := output.NewFile(*outputFileName)
	if err != nil {
		log.Fatalf("Failed to create output file: %s", *outputFileName)
	}
	defer outputFile.Close()

	// Create server for processing read command commands
	srv, err := server.NewServer(readQueue, writeQueue, *readWorkers, *writeWorkers, outputFile)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		cancel() // Cancel the context to stop all goroutines
	}()

	if err := srv.Start(ctx); err != nil {
		log.Fatalf("Server start error: %v", err)
	}

}
