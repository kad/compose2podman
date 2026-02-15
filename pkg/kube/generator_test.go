package kube

import (
	"strings"
	"testing"

	"github.com/kad/compose2podman/internal/types"
)

func TestKubeGenerator(t *testing.T) {
	compose := &types.ComposeFile{
		Version: "3",
		Services: map[string]types.Service{
			"redis": {
				Image: "redis:alpine",
				Ports: []string{"6379:6379"},
			},
		},
	}

	gen := NewGenerator(compose, "test-pod")
	yaml, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(yaml, "name: test-pod") {
		t.Error("Generated YAML should contain pod name")
	}

	if !strings.Contains(yaml, "image: redis:alpine") {
		t.Error("Generated YAML should contain image")
	}

	if !strings.Contains(yaml, "containerPort: 6379") {
		t.Error("Generated YAML should contain port")
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		input         string
		containerPort string
		hostPort      string
	}{
		{"8080:80", "80", "8080"},
		{"3000", "3000", ""},
		{"127.0.0.1:8080:80", "80", "127.0.0.1:8080"},
	}

	for _, tt := range tests {
		container, host := parsePort(tt.input)
		if container != tt.containerPort {
			t.Errorf("parsePort(%s): expected containerPort '%s', got '%s'", tt.input, tt.containerPort, container)
		}
		if host != tt.hostPort {
			t.Errorf("parsePort(%s): expected hostPort '%s', got '%s'", tt.input, tt.hostPort, host)
		}
	}
}

func TestGenerateWithEnvironment(t *testing.T) {
	compose := &types.ComposeFile{
		Services: map[string]types.Service{
			"app": {
				Image: "myapp:latest",
				Environment: map[string]interface{}{
					"NODE_ENV": "production",
					"PORT":     "3000",
				},
			},
		},
	}

	gen := NewGenerator(compose, "test-pod")
	yaml, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(yaml, "NODE_ENV") {
		t.Error("Generated YAML should contain environment variable NODE_ENV")
	}

	if !strings.Contains(yaml, "production") {
		t.Error("Generated YAML should contain environment value 'production'")
	}
}
