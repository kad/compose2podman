// Package kube provides functionality to generate Kubernetes Pod YAML files
// from Docker Compose definitions.
package kube

import (
	"fmt"
	"strings"

	"github.com/kad/compose2podman/internal/types"
)

// Generator generates Kubernetes YAML for podman play kube
type Generator struct {
	compose *types.ComposeFile
	podName string
}

// NewGenerator creates a new Kubernetes YAML generator
func NewGenerator(compose *types.ComposeFile, podName string) *Generator {
	if podName == "" {
		podName = "compose-pod"
	}
	return &Generator{
		compose: compose,
		podName: podName,
	}
}

// Generate creates Kubernetes Pod YAML
func (g *Generator) Generate() (string, error) {
	var sb strings.Builder

	sb.WriteString("apiVersion: v1\n")
	sb.WriteString("kind: Pod\n")
	sb.WriteString("metadata:\n")
	sb.WriteString(fmt.Sprintf("  name: %s\n", g.podName))

	// Add labels
	sb.WriteString("  labels:\n")
	sb.WriteString("    app: compose2podman\n")

	sb.WriteString("spec:\n")
	sb.WriteString("  containers:\n")

	// Generate containers from services
	for name, service := range g.compose.Services {
		if err := g.generateContainer(&sb, name, service); err != nil {
			return "", err
		}
	}

	// Add volumes
	if len(g.compose.Volumes) > 0 {
		sb.WriteString("  volumes:\n")
		for volName := range g.compose.Volumes {
			sb.WriteString(fmt.Sprintf("  - name: %s\n", volName))
			sb.WriteString("    persistentVolumeClaim:\n")
			sb.WriteString(fmt.Sprintf("      claimName: %s\n", volName))
		}
	}

	// Add restart policy
	sb.WriteString("  restartPolicy: Always\n")

	return sb.String(), nil
}

func (g *Generator) generateContainer(sb *strings.Builder, name string, service types.Service) error {
	containerName := name
	if service.ContainerName != "" {
		containerName = service.ContainerName
	}

	fmt.Fprintf(sb, "  - name: %s\n", containerName)

	if service.Image != "" {
		fmt.Fprintf(sb, "    image: %s\n", service.Image)
	} else {
		return fmt.Errorf("service %s: image is required (build not supported)", name)
	}

	// Command
	if cmd := service.CommandList(); len(cmd) > 0 {
		sb.WriteString("    command:\n")
		for _, c := range cmd {
			fmt.Fprintf(sb, "    - %s\n", c)
		}
	}

	// Entrypoint (args in K8s)
	if ep := service.EntrypointList(); len(ep) > 0 {
		sb.WriteString("    args:\n")
		for _, arg := range ep {
			fmt.Fprintf(sb, "    - %s\n", arg)
		}
	}

	// Environment variables
	env := service.EnvironmentMap()
	if len(env) > 0 {
		sb.WriteString("    env:\n")
		for key, val := range env {
			fmt.Fprintf(sb, "    - name: %s\n", key)
			fmt.Fprintf(sb, "      value: \"%s\"\n", val)
		}
	}

	// Ports
	if len(service.Ports) > 0 {
		sb.WriteString("    ports:\n")
		for _, port := range service.Ports {
			containerPort, hostPort := parsePort(port)
			fmt.Fprintf(sb, "    - containerPort: %s\n", containerPort)
			if hostPort != "" {
				fmt.Fprintf(sb, "      hostPort: %s\n", hostPort)
			}
		}
	}

	// Volume mounts
	if len(service.Volumes) > 0 {
		sb.WriteString("    volumeMounts:\n")
		for _, vol := range service.Volumes {
			mountPath, volumeName := parseVolume(vol)
			fmt.Fprintf(sb, "    - name: %s\n", volumeName)
			fmt.Fprintf(sb, "      mountPath: %s\n", mountPath)
		}
	}

	// Working directory
	if service.WorkingDir != "" {
		fmt.Fprintf(sb, "    workingDir: %s\n", service.WorkingDir)
	}

	// Security context
	if service.User != "" || service.Privileged {
		sb.WriteString("    securityContext:\n")
		if service.User != "" {
			// Parse user:group format
			uid, gid := parseUser(service.User)
			if uid != "" {
				fmt.Fprintf(sb, "      runAsUser: %s\n", uid)
			}
			if gid != "" {
				fmt.Fprintf(sb, "      runAsGroup: %s\n", gid)
			}
		}
		if service.Privileged {
			sb.WriteString("      privileged: true\n")
		}
	}

	return nil
}

// parsePort parses Docker Compose port format (host:container, ip:host:container or container)
func parsePort(port string) (containerPort, hostPort string) {
	parts := strings.Split(port, ":")
	switch len(parts) {
	case 3:
		// ip:host:container format
		return parts[2], parts[0] + ":" + parts[1]
	case 2:
		// host:container format
		return parts[1], parts[0]
	default:
		// container only
		return port, ""
	}
}

// parseVolume parses Docker Compose volume format (host:container or named:container)
func parseVolume(vol string) (mountPath, volumeName string) {
	parts := strings.Split(vol, ":")
	if len(parts) >= 2 {
		volumeName = strings.ReplaceAll(parts[0], "/", "-")
		volumeName = strings.ReplaceAll(volumeName, "_", "-")
		volumeName = strings.ToLower(volumeName)
		if volumeName[0] == '-' {
			volumeName = "vol" + volumeName
		}
		return parts[1], volumeName
	}
	return vol, "unnamed-vol"
}

// parseUser parses user:group format
func parseUser(user string) (uid, gid string) {
	parts := strings.Split(user, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return user, ""
}
