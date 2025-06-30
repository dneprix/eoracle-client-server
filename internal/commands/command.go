package commands

import (
	"encoding/json"
	"fmt"
)

type CommandType string
type CommandCategory string

type Command interface {
	ToJSON() ([]byte, error)
	GetType() CommandType
	GetKey() string
	GetValue() string
}

type command struct {
	Type  CommandType `json:"type"`
	Key   string      `json:"key,omitempty"`
	Value string      `json:"value,omitempty"`
}

// ToJSON converts command to JSON
func (c *command) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

// GetType returns the key of the command
func (c *command) GetType() CommandType {
	return c.Type
}

// GetKey returns the key of the command
func (c *command) GetKey() string {
	return c.Key
}

// GetValue returns the value of the command
func (c *command) GetValue() string {
	return c.Value
}

// FromJSON creates command from JSON
func FromJSON(data []byte) (Command, error) {
	var cmd command
	err := json.Unmarshal(data, &cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal command: %w", err)
	}
	return &cmd, nil
}
