package output

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestNewFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()
	outputFileName := filepath.Join(tempDir, "testfile.txt")

	// Test successful file creation
	f, err := NewFile(outputFileName)
	if err != nil {
		t.Fatalf("NewFile() returned error: %v", err)
	}
	if f == nil {
		t.Fatal("NewFile() returned nil File")
	}

	// Verify file exists
	if _, err := os.Stat(outputFileName); os.IsNotExist(err) {
		t.Errorf("Output file was not created")
	}

	// Test creation with invalid path
	invalidPath := "/invalid/path/testfile.txt"
	_, err = NewFile(invalidPath)
	if err == nil {
		t.Error("Expected error when creating file with invalid path, got nil")
	}
}

func TestWrite(t *testing.T) {
	tempDir := t.TempDir()
	outputFileName := filepath.Join(tempDir, "testfile.txt")

	f, err := NewFile(outputFileName)
	if err != nil {
		t.Fatalf("NewFile() returned error: %v", err)
	}
	defer f.Close()

	// Test writing data
	testData := "Hello, World!\n"
	f.Write(testData)

	// Read and verify file contents
	content, err := os.ReadFile(outputFileName)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if string(content) != testData {
		t.Errorf("Expected file content %q, got %q", testData, string(content))
	}
}

func TestWriteConcurrent(t *testing.T) {
	tempDir := t.TempDir()
	outputFileName := filepath.Join(tempDir, "testfile.txt")

	f, err := NewFile(outputFileName)
	if err != nil {
		t.Fatalf("NewFile() returned error: %v", err)
	}
	defer f.Close()

	// Test concurrent writes
	var wg sync.WaitGroup
	numGoroutines := 100
	dataToWrite := "Test data\n"
	expectedContent := ""

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.Write(dataToWrite)
		}()
		expectedContent += dataToWrite
	}
	wg.Wait()

	// Read and verify file contents
	content, err := os.ReadFile(outputFileName)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Check if all writes were successful (content length should match)
	if len(content) != len(expectedContent) {
		t.Errorf("Expected content length %d, got %d", len(expectedContent), len(content))
	}
}

func TestClose(t *testing.T) {
	tempDir := t.TempDir()
	outputFileName := filepath.Join(tempDir, "testfile.txt")

	f, err := NewFile(outputFileName)
	if err != nil {
		t.Fatalf("NewFile() returned error: %v", err)
	}

	// Test closing file
	err = f.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}

}
