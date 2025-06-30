package commands

import (
	"strings"
	"testing"
)

func TestCommand_ToJSON(t *testing.T) {
	tests := []struct {
		name    string
		command command
		want    string
		wantErr bool
	}{
		{
			name: "complete command",
			command: command{
				Type:  "SET",
				Key:   "user:123",
				Value: "john_doe",
			},
			want:    `{"type":"SET","key":"user:123","value":"john_doe"}`,
			wantErr: false,
		},
		{
			name: "command with only type",
			command: command{
				Type: "PING",
			},
			want:    `{"type":"PING"}`,
			wantErr: false,
		},
		{
			name: "command with type and key",
			command: command{
				Type: "GET",
				Key:  "config:timeout",
			},
			want:    `{"type":"GET","key":"config:timeout"}`,
			wantErr: false,
		},
		{
			name:    "empty command",
			command: command{},
			want:    `{"type":""}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.command.ToJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("Command.ToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("Command.ToJSON() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestFromJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    *command
		wantErr bool
	}{
		{
			name: "valid complete JSON",
			data: []byte(`{"type":"SET","key":"user:123","value":"john_doe"}`),
			want: &command{
				Type:  "SET",
				Key:   "user:123",
				Value: "john_doe",
			},
			wantErr: false,
		},
		{
			name: "valid minimal JSON",
			data: []byte(`{"type":"PING"}`),
			want: &command{
				Type: "PING",
			},
			wantErr: false,
		},
		{
			name: "JSON with extra fields (should be ignored)",
			data: []byte(`{"type":"GET","key":"config","extra":"ignored"}`),
			want: &command{
				Type: "GET",
				Key:  "config",
			},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			data:    []byte(`{"type":"SET","key":}`),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty JSON object",
			data:    []byte(`{}`),
			want:    &command{},
			wantErr: false,
		},
		{
			name:    "null JSON",
			data:    []byte(`null`),
			want:    &command{},
			wantErr: false,
		},
		{
			name:    "empty byte slice",
			data:    []byte(``),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromJSON(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !commandsEqual(got, tt.want) {
				t.Errorf("FromJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromJSON_ErrorMessage(t *testing.T) {
	invalidJSON := []byte(`{"type":"SET","key":}`)
	_, err := FromJSON(invalidJSON)

	if err == nil {
		t.Error("FromJSON() expected error for invalid JSON")
		return
	}

	if !strings.Contains(err.Error(), "failed to unmarshal command") {
		t.Errorf("FromJSON() error message should contain 'failed to unmarshal command', got: %v", err.Error())
	}
}

func TestJSONRoundTrip(t *testing.T) {
	original := command{
		Type:  "UPDATE",
		Key:   "settings:theme",
		Value: "dark",
	}

	// Convert to JSON
	jsonData, err := original.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() failed: %v", err)
	}

	// Convert back from JSON
	restored, err := FromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON() failed: %v", err)
	}

	// Compare
	if !commandsEqual(&original, restored) {
		t.Errorf("Round trip failed: original = %v, restored = %v", original, restored)
	}
}

func TestJSONOmitEmpty(t *testing.T) {
	cmd := command{
		Type: "STATUS",
		// Key and Value are empty, should be omitted from JSON
	}

	jsonData, err := cmd.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() failed: %v", err)
	}

	jsonStr := string(jsonData)
	if strings.Contains(jsonStr, "key") {
		t.Error("JSON should not contain 'key' field when empty")
	}
	if strings.Contains(jsonStr, "value") {
		t.Error("JSON should not contain 'value' field when empty")
	}
	if !strings.Contains(jsonStr, `"type":"STATUS"`) {
		t.Error("JSON should contain type field")
	}
}

func TestCommandTypes(t *testing.T) {
	var cmdType CommandType = "CUSTOM_TYPE"
	var category CommandCategory = "CUSTOM_CATEGORY"

	cmd := command{
		Type: cmdType,
	}

	// Test that custom types work
	jsonData, err := cmd.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() failed with custom type: %v", err)
	}

	restored, err := FromJSON(jsonData)
	if err != nil {
		t.Fatalf("FromJSON() failed with custom type: %v", err)
	}

	if restored.GetType() != cmdType {
		t.Errorf("Custom CommandType not preserved: got %v, want %v", restored.GetType(), cmdType)
	}

	// Just verify CommandCategory type exists and can be used
	_ = category
}

// Helper function to compare commands
func commandsEqual(a, b Command) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.GetType() == b.GetType() && a.GetKey() == b.GetKey() && a.GetValue() == b.GetValue()
}

// Benchmark tests
func BenchmarkCommand_ToJSON(b *testing.B) {
	cmd := command{
		Type:  "BENCHMARK",
		Key:   "test:key:12345",
		Value: "benchmark_value_with_some_length",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cmd.ToJSON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFromJSON(b *testing.B) {
	jsonData := []byte(`{"type":"BENCHMARK","key":"test:key:12345","value":"benchmark_value_with_some_length"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := FromJSON(jsonData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONRoundTrip(b *testing.B) {
	cmd := command{
		Type:  "BENCHMARK",
		Key:   "test:key:12345",
		Value: "benchmark_value_with_some_length",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jsonData, err := cmd.ToJSON()
		if err != nil {
			b.Fatal(err)
		}
		_, err = FromJSON(jsonData)
		if err != nil {
			b.Fatal(err)
		}
	}
}
