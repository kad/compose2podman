package parser

import (
	"os"
	"testing"
)

func TestParseComposeFile(t *testing.T) {
	// Create a temporary test file
	content := []byte(`version: "3"
services:
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
`)
	tmpfile, err := os.CreateTemp("", "compose-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if removeErr := os.Remove(tmpfile.Name()); removeErr != nil {
			t.Logf("Failed to remove temp file: %v", removeErr)
		}
	}()

	if _, writeErr := tmpfile.Write(content); writeErr != nil {
		t.Fatal(writeErr)
	}
	if closeErr := tmpfile.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}

	// Test parsing
	compose, err := ParseComposeFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("ParseComposeFile failed: %v", err)
	}

	if compose.Version != "3" {
		t.Errorf("Expected version '3', got '%s'", compose.Version)
	}

	if len(compose.Services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(compose.Services))
	}

	redis, ok := compose.Services["redis"]
	if !ok {
		t.Fatal("Expected 'redis' service not found")
	}

	if redis.Image != "redis:alpine" {
		t.Errorf("Expected image 'redis:alpine', got '%s'", redis.Image)
	}

	if len(redis.Ports) != 1 {
		t.Errorf("Expected 1 port, got %d", len(redis.Ports))
	}
}

func TestParseComposeFileNotFound(t *testing.T) {
	_, err := ParseComposeFile("nonexistent-file.yaml")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestParseComposeFileInvalidYAML(t *testing.T) {
	// Create a temporary test file with invalid YAML
	content := []byte(`invalid: yaml: content: [`)
	tmpfile, err := os.CreateTemp("", "compose-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if removeErr := os.Remove(tmpfile.Name()); removeErr != nil {
			t.Logf("Failed to remove temp file: %v", removeErr)
		}
	}()

	if _, writeErr := tmpfile.Write(content); writeErr != nil {
		t.Fatal(writeErr)
	}
	if closeErr := tmpfile.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}

	_, err = ParseComposeFile(tmpfile.Name())
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}
