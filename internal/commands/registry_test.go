package commands

import (
	"eoracle-client-server/internal/storage"
	"strings"
	"testing"
)

type mockStorage struct {
	data map[string]string
}

func (m *mockStorage) Add(key, value string) {
	m.data[key] = value
}

func (m *mockStorage) Delete(key string) bool {
	if _, exists := m.data[key]; exists {
		delete(m.data, key)
		return true
	}
	return false
}

func (m *mockStorage) Get(key string) (string, bool) {
	value, exists := m.data[key]
	return value, exists
}

func (m *mockStorage) GetAll() []storage.KeyValue {
	var items []storage.KeyValue
	for k, v := range m.data {
		items = append(items, storage.KeyValue{Key: k, Value: v})
	}
	return items
}

type mockOutput struct {
	output strings.Builder
}

func (m *mockOutput) Write(s string) {
	m.output.WriteString(s)
}
func (m *mockOutput) Close() error {
	return nil
}

func TestCommandRegistry_ParseCommand(t *testing.T) {
	registry := NewCommandRegistry()

	tests := []struct {
		name        string
		input       string
		wantErr     bool
		wantCommand Command
	}{
		{
			name:    "valid add command",
			input:   "add key value",
			wantErr: false,
			wantCommand: &command{
				Type:  AddItem,
				Key:   "key",
				Value: "value",
			},
		},
		{
			name:    "invalid add command - missing value",
			input:   "add key",
			wantErr: true,
		},
		{
			name:    "valid delete command",
			input:   "delete key",
			wantErr: false,
			wantCommand: &command{
				Type: DeleteItem,
				Key:  "key",
			},
		},
		{
			name:    "invalid delete command - missing key",
			input:   "delete",
			wantErr: true,
		},
		{
			name:    "valid get command",
			input:   "get key",
			wantErr: false,
			wantCommand: &command{
				Type: GetItem,
				Key:  "key",
			},
		},
		{
			name:    "valid getall command",
			input:   "getall",
			wantErr: false,
			wantCommand: &command{
				Type: GetAllItems,
			},
		},
		{
			name:    "unknown command",
			input:   "invalid cmd",
			wantErr: true,
		},
		{
			name:    "empty command",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := registry.ParseCommand(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && cmd != nil {
				if cmd.GetType() != tt.wantCommand.GetType() ||
					cmd.GetKey() != tt.wantCommand.GetKey() ||
					cmd.GetValue() != tt.wantCommand.GetValue() {
					t.Errorf("ParseCommand() = %v, want %v", cmd, tt.wantCommand)
				}
			}
		})
	}
}

func TestCommandRegistry_HandleCommand(t *testing.T) {
	registry := NewCommandRegistry()
	store := &mockStorage{data: make(map[string]string)}
	out := &mockOutput{}

	tests := []struct {
		name       string
		command    Command
		setup      func()
		wantErr    bool
		wantOutput string
	}{
		{
			name: "add command",
			command: &command{
				Type:  AddItem,
				Key:   "key1",
				Value: "value1",
			},
			setup:   func() {},
			wantErr: false,
		},
		{
			name: "delete existing key",
			command: &command{
				Type: DeleteItem,
				Key:  "key1",
			},
			setup: func() {
				store.Add("key1", "value1")
			},
			wantErr: false,
		},
		{
			name: "get existing key",
			command: &command{
				Type: GetItem,
				Key:  "key1",
			},
			setup: func() {
				store.Add("key1", "value1")
			},
			wantErr:    false,
			wantOutput: "key1 = value1\n",
		},
		{
			name: "getall with items",
			command: &command{
				Type: GetAllItems,
			},
			setup: func() {
				store.Add("key1", "value1")
				store.Add("key2", "value2")
			},
			wantErr:    false,
			wantOutput: "key1 = value1\nkey2 = value2\n",
		},
		{
			name: "unknown command type",
			command: &command{
				Type: "unknown",
			},
			setup:   func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset storage and output
			store.data = make(map[string]string)
			out.output = strings.Builder{}

			tt.setup()
			err := registry.HandleCommand(tt.command, store, out)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantOutput != "" && out.output.String() != tt.wantOutput {
				t.Errorf("HandleCommand() output = %v, want %v", out.output.String(), tt.wantOutput)
			}
		})
	}
}

func TestCommandRegistry_IsReadCommand(t *testing.T) {
	registry := NewCommandRegistry()

	tests := []struct {
		name    string
		command Command
		want    bool
	}{
		{
			name:    "get command",
			command: &command{Type: GetItem},
			want:    true,
		},
		{
			name:    "getall command",
			command: &command{Type: GetAllItems},
			want:    true,
		},
		{
			name:    "add command",
			command: &command{Type: AddItem},
			want:    false,
		},
		{
			name:    "delete command",
			command: &command{Type: DeleteItem},
			want:    false,
		},
		{
			name:    "unknown command",
			command: &command{Type: "unknown"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := registry.IsReadCommand(tt.command); got != tt.want {
				t.Errorf("IsReadCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
