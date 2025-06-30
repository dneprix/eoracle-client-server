package storage

import (
	"sync"
)

// OrderedMap represents a thread-safe ordered map with O(1) operations
type OrderedMap struct {
	mu   sync.RWMutex
	data map[string]*node
	head *node
	tail *node
	size int
}

type node struct {
	key   string
	value string
	next  *node
	prev  *node
}

// KeyValue represents a key-value pair
type KeyValue struct {
	Key   string
	Value string
}

// New creates a new OrderedMap
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		data: make(map[string]*node),
	}
}

// Add adds or updates a key-value pair in O(1) time
func (om *OrderedMap) Add(key, value string) {
	om.mu.Lock()
	defer om.mu.Unlock()

	if existingNode, exists := om.data[key]; exists {
		// Update existing node
		existingNode.value = value
		return
	}

	// Create new node
	newNode := &node{
		key:   key,
		value: value,
	}

	// Add to map
	om.data[key] = newNode

	// Add to linked list
	if om.head == nil {
		om.head = newNode
		om.tail = newNode
	} else {
		om.tail.next = newNode
		newNode.prev = om.tail
		om.tail = newNode
	}

	om.size++
}

// Get retrieves a value by key in O(1) time
func (om *OrderedMap) Get(key string) (string, bool) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	if node, exists := om.data[key]; exists {
		return node.value, true
	}
	return "", false
}

// Delete removes a key-value pair in O(1) time
func (om *OrderedMap) Delete(key string) bool {
	om.mu.Lock()
	defer om.mu.Unlock()

	node, exists := om.data[key]
	if !exists {
		return false
	}

	// Remove from map
	delete(om.data, key)

	// Remove from linked list
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		om.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		om.tail = node.prev
	}

	om.size--
	return true
}

// GetAll returns all key-value pairs in insertion order
func (om *OrderedMap) GetAll() []KeyValue {
	om.mu.RLock()
	defer om.mu.RUnlock()

	result := make([]KeyValue, 0, om.size)
	current := om.head

	for current != nil {
		result = append(result, KeyValue{
			Key:   current.key,
			Value: current.value,
		})
		current = current.next
	}

	return result
}

// Size returns the number of elements
func (om *OrderedMap) Size() int {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.size
}
