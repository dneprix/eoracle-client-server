package storage

import (
	"reflect"
	"testing"
)

// TestNewOrderedMap tests the creation of a new OrderedMap
func TestNewOrderedMap(t *testing.T) {
	om := NewOrderedMap()
	if om == nil {
		t.Fatal("NewOrderedMap returned nil")
	}
	if om.data == nil {
		t.Fatal("NewOrderedMap data map is nil")
	}
	if om.size != 0 {
		t.Errorf("Expected size 0, got %d", om.size)
	}
	if om.head != nil || om.tail != nil {
		t.Error("Expected head and tail to be nil")
	}
}

// TestAdd tests adding new key-value pairs and updating existing ones
func TestAdd(t *testing.T) {
	om := NewOrderedMap()

	// Test adding new key-value pair
	om.Add("key1", "value1")
	if om.Size() != 1 {
		t.Errorf("Expected size 1, got %d", om.Size())
	}
	if val, exists := om.Get("key1"); !exists || val != "value1" {
		t.Errorf("Expected key1=value1, got exists=%v, value=%s", exists, val)
	}

	// Test updating existing key
	om.Add("key1", "value2")
	if om.Size() != 1 {
		t.Errorf("Expected size 1 after update, got %d", om.Size())
	}
	if val, exists := om.Get("key1"); !exists || val != "value2" {
		t.Errorf("Expected key1=value2, got exists=%v, value=%s", exists, val)
	}

	// Test adding multiple keys
	om.Add("key2", "value2")
	if om.Size() != 2 {
		t.Errorf("Expected size 2, got %d", om.Size())
	}
}

// TestGet tests retrieving values by key
func TestGet(t *testing.T) {
	om := NewOrderedMap()

	// Test getting non-existent key
	if val, exists := om.Get("key1"); exists || val != "" {
		t.Errorf("Expected non-existent key to return false and empty string, got exists=%v, value=%s", exists, val)
	}

	// Test getting existing key
	om.Add("key1", "value1")
	if val, exists := om.Get("key1"); !exists || val != "value1" {
		t.Errorf("Expected key1=value1, got exists=%v, value=%s", exists, val)
	}
}

// TestDelete tests removing key-value pairs
func TestDelete(t *testing.T) {
	om := NewOrderedMap()

	// Test deleting non-existent key
	if om.Delete("key1") {
		t.Error("Expected Delete to return false for non-existent key")
	}

	// Test deleting existing keys
	om.Add("key1", "value1")
	om.Add("key2", "value2")
	om.Add("key3", "value3")

	// Delete middle element
	if !om.Delete("key2") {
		t.Error("Expected Delete to return true for existing key")
	}
	if om.Size() != 2 {
		t.Errorf("Expected size 2 after delete, got %d", om.Size())
	}

	// Delete head
	if !om.Delete("key1") {
		t.Error("Expected Delete to return true for head")
	}
	if om.Size() != 1 {
		t.Errorf("Expected size 1 after delete, got %d", om.Size())
	}

	// Delete tail
	if !om.Delete("key3") {
		t.Error("Expected Delete to return true for tail")
	}
	if om.Size() != 0 {
		t.Errorf("Expected size 0 after delete, got %d", om.Size())
	}
}

// TestGetAll tests retrieving all key-value pairs in insertion order
func TestGetAll(t *testing.T) {
	om := NewOrderedMap()

	// Test empty map
	if len(om.GetAll()) != 0 {
		t.Errorf("Expected empty slice for empty map, got %v", om.GetAll())
	}

	// Test multiple elements
	om.Add("key1", "value1")
	om.Add("key2", "value2")
	om.Add("key3", "value3")

	expected := []KeyValue{
		{Key: "key1", Value: "value1"},
		{Key: "key2", Value: "value2"},
		{Key: "key3", Value: "value3"},
	}

	result := om.GetAll()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// TestSize tests the size tracking functionality
func TestSize(t *testing.T) {
	om := NewOrderedMap()

	if om.Size() != 0 {
		t.Errorf("Expected size 0, got %d", om.Size())
	}

	om.Add("key1", "value1")
	om.Add("key2", "value2")
	if om.Size() != 2 {
		t.Errorf("Expected size 2, got %d", om.Size())
	}

	om.Delete("key1")
	if om.Size() != 1 {
		t.Errorf("Expected size 1 after delete, got %d", om.Size())
	}
}

// TestConcurrentAccess tests thread-safety with concurrent operations
func TestConcurrentAccess(t *testing.T) {
	om := NewOrderedMap()
	done := make(chan bool)

	// Concurrent adds
	go func() {
		for i := 0; i < 100; i++ {
			om.Add("key"+string(rune(i)), "value"+string(rune(i)))
		}
		done <- true
	}()

	// Concurrent gets
	go func() {
		for i := 0; i < 100; i++ {
			om.Get("key" + string(rune(i)))
		}
		done <- true
	}()

	// Concurrent deletes
	go func() {
		for i := 0; i < 50; i++ {
			om.Delete("key" + string(rune(i)))
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify final size
	if om.Size() > 100 || om.Size() < 50 {
		t.Errorf("Unexpected size after concurrent operations: %d", om.Size())
	}
}
