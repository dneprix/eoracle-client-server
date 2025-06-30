package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strings"

	"eoracle-client-server/internal/commands"
	"eoracle-client-server/internal/queue"
)

func main() {
	var (
		rabbitURL      = flag.String("rabbit-url", "amqp://guest:guest@localhost:5672/", "RabbitMQ URL")
		readQueueName  = flag.String("read-queue-name", "read-commands", "Queue name for read commands")
		writeQueueName = flag.String("write-queue-name", "write-commands", "Queue name for write commands")
	)
	flag.Parse()

	log.Printf("Starting client")

	// Create read queue
	readQueue, err := queue.NewRabbitMQQueue(*rabbitURL, *readQueueName)
	if err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}
	defer readQueue.Close()

	// Create write queue
	writeQueue, err := queue.NewRabbitMQQueue(*rabbitURL, *writeQueueName)
	if err != nil {
		log.Fatalf("Failed to create queue: %v", err)
	}
	defer writeQueue.Close()

	commandRegistry := commands.NewCommandRegistry()

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		// Read a line from stdin
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse the command
		cmd, err := commandRegistry.ParseCommand(line)
		if err != nil {
			log.Printf("Invalid command: %v", err)
			continue
		}

		// Check if the command is read or write
		if commandRegistry.IsReadCommand(cmd) {
			if err := readQueue.Publish(cmd); err != nil {
				log.Printf("Failed to publish read command: %v", err)
				continue
			}
		} else {
			if err := writeQueue.Publish(cmd); err != nil {
				log.Printf("Failed to publish write command: %v", err)
				continue
			}
		}
		log.Printf("Published command: %s", line)

	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Scanner error: %v", err)
	}
}
