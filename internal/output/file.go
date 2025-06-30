package output

import (
	"log"
	"os"
	"sync"
)

type file struct {
	outputFile *os.File
	mu         sync.Mutex // Mutex to ensure thread-safe writes
}

func NewFile(outputFileName string) (Output, error) {
	// Open the output file for writing, creating it if it doesn't exist
	outputFile, err := os.OpenFile(outputFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &file{
		outputFile: outputFile,
	}, nil
}

// writeToFile writes data to the output file in a thread-safe manner
func (f *file) Write(data string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, err := f.outputFile.WriteString(data); err != nil {
		log.Printf("Failed to write to file: %v", err)
	}
	f.outputFile.Sync()
}

func (f *file) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.outputFile.Close(); err != nil {
		return err
	}
	return nil
}
