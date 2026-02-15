package quadlet

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kad/compose2podman/internal/types"
)

// Generator generates Podman Quadlet files
type Generator struct {
	compose   *types.ComposeFile
	outputDir string
}

// NewGenerator creates a new Quadlet generator
func NewGenerator(compose *types.ComposeFile, outputDir string) *Generator {
	return &Generator{
		compose:   compose,
		outputDir: outputDir,
	}
}

// Generate creates Quadlet files (.container, .volume, .network)
func (g *Generator) Generate() error {
	// Create output directory
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate network files
	for name, network := range g.compose.Networks {
		if err := g.generateNetwork(name, network); err != nil {
			return err
		}
	}

	// Generate volume files
	for name, volume := range g.compose.Volumes {
		if err := g.generateVolume(name, volume); err != nil {
			return err
		}
	}

	// Generate container files
	for name, service := range g.compose.Services {
		if err := g.generateContainer(name, service); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) generateContainer(name string, service types.Service) error {
	var sb strings.Builder

	sb.WriteString("[Unit]\n")
	sb.WriteString(fmt.Sprintf("Description=%s container\n", name))

	// Handle dependencies
	deps := service.DependsOnList()
	if len(deps) > 0 {
		var after []string
		for _, dep := range deps {
			after = append(after, fmt.Sprintf("%s.service", dep))
		}
		sb.WriteString(fmt.Sprintf("After=%s\n", strings.Join(after, " ")))
		sb.WriteString(fmt.Sprintf("Requires=%s\n", strings.Join(after, " ")))
	}

	sb.WriteString("\n[Container]\n")

	if service.Image != "" {
		sb.WriteString(fmt.Sprintf("Image=%s\n", service.Image))
	}

	// Container name
	containerName := name
	if service.ContainerName != "" {
		containerName = service.ContainerName
	}
	sb.WriteString(fmt.Sprintf("ContainerName=%s\n", containerName))

	// Environment variables
	env := service.EnvironmentMap()
	for key, val := range env {
		sb.WriteString(fmt.Sprintf("Environment=%s=%s\n", key, val))
	}

	// Ports
	for _, port := range service.Ports {
		sb.WriteString(fmt.Sprintf("PublishPort=%s\n", port))
	}

	// Volumes
	for _, vol := range service.Volumes {
		sb.WriteString(fmt.Sprintf("Volume=%s\n", vol))
	}

	// Networks
	networks := service.NetworksList()
	for _, net := range networks {
		sb.WriteString(fmt.Sprintf("Network=%s.network\n", net))
	}

	// Working directory
	if service.WorkingDir != "" {
		sb.WriteString(fmt.Sprintf("WorkingDir=%s\n", service.WorkingDir))
	}

	// User
	if service.User != "" {
		sb.WriteString(fmt.Sprintf("User=%s\n", service.User))
	}

	// Command
	if cmd := service.CommandList(); len(cmd) > 0 {
		sb.WriteString(fmt.Sprintf("Exec=%s\n", strings.Join(cmd, " ")))
	}

	// Hostname
	if service.Hostname != "" {
		sb.WriteString(fmt.Sprintf("HostName=%s\n", service.Hostname))
	}

	// Privileged
	if service.Privileged {
		sb.WriteString("SecurityLabelDisable=true\n")
	}

	// Capabilities
	for _, cap := range service.CapAdd {
		sb.WriteString(fmt.Sprintf("AddCapability=%s\n", cap))
	}
	for _, cap := range service.CapDrop {
		sb.WriteString(fmt.Sprintf("DropCapability=%s\n", cap))
	}

	// Labels
	for key, val := range service.Labels {
		sb.WriteString(fmt.Sprintf("Label=%s=%s\n", key, val))
	}

	sb.WriteString("\n[Service]\n")

	// Map Docker Compose restart policies to systemd
	restart := "always"
	switch service.Restart {
	case "no":
		restart = "no"
	case "always":
		restart = "always"
	case "on-failure":
		restart = "on-failure"
	case "unless-stopped":
		restart = "always"
	}
	sb.WriteString(fmt.Sprintf("Restart=%s\n", restart))
	sb.WriteString("TimeoutStartSec=900\n")

	sb.WriteString("\n[Install]\n")
	sb.WriteString("WantedBy=default.target\n")

	// Write file
	filename := filepath.Join(g.outputDir, fmt.Sprintf("%s.container", name))
	if err := os.WriteFile(filename, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write container file: %w", err)
	}

	return nil
}

func (g *Generator) generateVolume(name string, volume types.Volume) error {
	var sb strings.Builder

	sb.WriteString("[Unit]\n")
	sb.WriteString(fmt.Sprintf("Description=%s volume\n", name))

	sb.WriteString("\n[Volume]\n")

	if volume.Driver != "" && volume.Driver != "local" {
		sb.WriteString(fmt.Sprintf("Driver=%s\n", volume.Driver))
	}

	// Labels
	for key, val := range volume.Labels {
		sb.WriteString(fmt.Sprintf("Label=%s=%s\n", key, val))
	}

	sb.WriteString("\n[Install]\n")
	sb.WriteString("WantedBy=default.target\n")

	// Write file
	filename := filepath.Join(g.outputDir, fmt.Sprintf("%s.volume", name))
	if err := os.WriteFile(filename, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write volume file: %w", err)
	}

	return nil
}

func (g *Generator) generateNetwork(name string, network types.Network) error {
	var sb strings.Builder

	sb.WriteString("[Unit]\n")
	sb.WriteString(fmt.Sprintf("Description=%s network\n", name))

	sb.WriteString("\n[Network]\n")

	if network.Driver != "" {
		sb.WriteString(fmt.Sprintf("Driver=%s\n", network.Driver))
	}

	// Labels
	for key, val := range network.Labels {
		sb.WriteString(fmt.Sprintf("Label=%s=%s\n", key, val))
	}

	sb.WriteString("\n[Install]\n")
	sb.WriteString("WantedBy=default.target\n")

	// Write file
	filename := filepath.Join(g.outputDir, fmt.Sprintf("%s.network", name))
	if err := os.WriteFile(filename, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write network file: %w", err)
	}

	return nil
}
