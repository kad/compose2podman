// Package parser provides functionality to parse Docker Compose YAML files.
package parser

import (
	"fmt"
	"os"

	"github.com/kad/compose2podman/internal/types"
	"gopkg.in/yaml.v3"
)

// ParseComposeFile reads and parses a Docker Compose file.
// The filename parameter is intentionally user-controlled for CLI tool functionality.
// nolint:gosec // G304: File path comes from CLI argument, expected behavior
func ParseComposeFile(filename string) (*types.ComposeFile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	var compose types.ComposeFile
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil, fmt.Errorf("failed to parse compose file: %w", err)
	}

	// Set default values
	if compose.Services == nil {
		compose.Services = make(map[string]types.Service)
	}
	if compose.Networks == nil {
		compose.Networks = make(map[string]types.Network)
	}
	if compose.Volumes == nil {
		compose.Volumes = make(map[string]types.Volume)
	}

	return &compose, nil
}
