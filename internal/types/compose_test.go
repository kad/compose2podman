package types

import "testing"

func TestServiceEnvironmentMap(t *testing.T) {
	tests := []struct {
		name     string
		env      interface{}
		expected map[string]string
	}{
		{
			name: "map string interface",
			env: map[string]interface{}{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name: "array format",
			env: []interface{}{
				"KEY1=value1",
				"KEY2=value2",
			},
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := Service{Environment: tt.env}
			result := svc.EnvironmentMap()

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d env vars, got %d", len(tt.expected), len(result))
			}

			for key, expectedVal := range tt.expected {
				if result[key] != expectedVal {
					t.Errorf("For key %s: expected '%s', got '%s'", key, expectedVal, result[key])
				}
			}
		})
	}
}

func TestServiceCommandList(t *testing.T) {
	tests := []struct {
		name     string
		command  interface{}
		expected []string
	}{
		{
			name:     "string command",
			command:  "node server.js",
			expected: []string{"node server.js"},
		},
		{
			name: "array command",
			command: []interface{}{
				"node",
				"server.js",
			},
			expected: []string{"node", "server.js"},
		},
		{
			name:     "nil command",
			command:  nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := Service{Command: tt.command}
			result := svc.CommandList()

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d items, got %d", len(tt.expected), len(result))
				return
			}

			for i, expected := range tt.expected {
				if result[i] != expected {
					t.Errorf("At index %d: expected '%s', got '%s'", i, expected, result[i])
				}
			}
		})
	}
}

func TestServiceNetworksList(t *testing.T) {
	tests := []struct {
		name     string
		networks interface{}
		expected []string
	}{
		{
			name: "array of strings",
			networks: []interface{}{
				"frontend",
				"backend",
			},
			expected: []string{"frontend", "backend"},
		},
		{
			name: "map format",
			networks: map[string]interface{}{
				"frontend": nil,
				"backend":  nil,
			},
			expected: []string{"frontend", "backend"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := Service{Networks: tt.networks}
			result := svc.NetworksList()

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d networks, got %d", len(tt.expected), len(result))
			}
		})
	}
}
