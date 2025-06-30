package server

import (
	"context"
	"log"
	"sync"

	"eoracle-client-server/internal/commands"
	"eoracle-client-server/internal/output"
	"eoracle-client-server/internal/queue"
	"eoracle-client-server/internal/storage"
)

type Server interface {
	Start(ctx context.Context) error
	Close() error
}

// Server represents the main server
type server struct {
	orderedMap   storage.Storage
	output       output.Output
	commands     commands.CommandRegistry
	readQueue    queue.Queue
	writeQueue   queue.Queue
	mu           sync.Mutex
	readWorkers  int
	writeWorkers int
}

// NewServer creates a new server
func NewServer(readQueue queue.Queue, writeQueue queue.Queue, readWorkers int, writeWorkers int, out output.Output) (Server, error) {

	return &server{
		orderedMap:   storage.NewOrderedMap(),
		commands:     commands.NewCommandRegistry(),
		readQueue:    readQueue,
		writeQueue:   writeQueue,
		output:       out,
		readWorkers:  readWorkers,
		writeWorkers: writeWorkers,
	}, nil
}

// Start starts the server
func (s *server) Start(ctx context.Context) error {
	log.Printf("Starting server with %d read workers and %d write workers", s.readWorkers, s.writeWorkers)

	// Create an unbuffered channels for read and write commands
	readCommandChan := make(chan commands.Command)
	writeCommandChan := make(chan commands.Command)

	// Use a WaitGroup to wait for all workers to finish
	var wg sync.WaitGroup

	// Start read worker goroutines
	for i := 0; i < s.readWorkers; i++ {
		wg.Add(1)
		go s.worker(ctx, &wg, readCommandChan)
	}

	// Start write worker goroutines
	for i := 0; i < s.writeWorkers; i++ {
		wg.Add(1)
		go s.worker(ctx, &wg, writeCommandChan)
	}

	// Start consuming messages from the read queue
	wg.Add(1)
	go func() {
		defer wg.Done()
		//Subscribe to the read queue
		if err := s.readQueue.Subscribe(ctx, func(cmd commands.Command) error {
			select {
			case readCommandChan <- cmd:
				return nil // If command is sent to channel, return nil
			case <-ctx.Done():
				return ctx.Err() // If context is done, return the error
			}
		}); err != nil {
			log.Fatalf("failed to subscribe to read queue: %s", err)
		}
	}()

	// Start consuming messages from the write queue
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Subscribe to the write queue
		if err := s.writeQueue.Subscribe(ctx, func(cmd commands.Command) error {

			select {
			case writeCommandChan <- cmd:
				return nil // If command is sent to channel, return nil
			case <-ctx.Done():
				return ctx.Err() // If context is done, return the error
			}
		}); err != nil {
			log.Fatalf("failed to subscribe to write queue: %s", err)
		}
	}()

	wg.Wait()
	log.Println("All goroutines have exited")
	return nil
}

// worker processes commands from the channel
func (s *server) worker(ctx context.Context, wg *sync.WaitGroup, commandChan <-chan commands.Command) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case cmd := <-commandChan:
			if err := s.commands.HandleCommand(cmd, s.orderedMap, s.output); err != nil {
				log.Printf("Failed to handle command: %v", err)
				continue
			}
		}
	}
}

// Close closes the server
func (s *server) Close() error {

	// Close the write queue
	if err := s.writeQueue.Close(); err != nil {
		log.Printf("Failed to close write queue: %v", err)
		return err
	}
	// Close the read queue
	if err := s.readQueue.Close(); err != nil {
		log.Printf("Failed to close read queue: %v", err)
	}
	return nil
}
