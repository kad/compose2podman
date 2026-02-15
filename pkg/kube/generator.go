// Package kube provides functionality to generate Kubernetes Pod YAML files
// from Docker Compose definitions.
package kube

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kad/compose2podman/internal/types"
)

// volumeInfo tracks information about volumes used in the pod
type volumeInfo struct {
	name     string
	hostPath string // empty if it's a named volume
	isPath   bool   // true if it's a host path (absolute or relative)
}

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

	// Track volumes used by containers
	usedVolumes := make(map[string]*volumeInfo)

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
		if err := g.generateContainer(&sb, name, service, usedVolumes); err != nil {
			return "", err
		}
	}

	// Add volumes section
	if len(usedVolumes) > 0 {
		sb.WriteString("  volumes:\n")
		for _, volInfo := range usedVolumes {
			sb.WriteString(fmt.Sprintf("  - name: %s\n", volInfo.name))
			if volInfo.isPath {
				// Use hostPath for actual paths
				sb.WriteString("    hostPath:\n")
				sb.WriteString(fmt.Sprintf("      path: %s\n", volInfo.hostPath))
				sb.WriteString("      type: DirectoryOrCreate\n")
			} else {
				// Use PVC for named volumes
				sb.WriteString("    persistentVolumeClaim:\n")
				sb.WriteString(fmt.Sprintf("      claimName: %s\n", volInfo.name))
			}
		}
	}

	// Add restart policy
	sb.WriteString("  restartPolicy: Always\n")

	return sb.String(), nil
}

func (g *Generator) generateContainer(sb *strings.Builder, name string, service types.Service, usedVolumes map[string]*volumeInfo) error {
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
			mountPath, volumeName, hostPath, isPath := parseVolume(vol)
			fmt.Fprintf(sb, "    - name: %s\n", volumeName)
			fmt.Fprintf(sb, "      mountPath: %s\n", mountPath)

			// Track volume for volumes section
			if _, exists := usedVolumes[volumeName]; !exists {
				usedVolumes[volumeName] = &volumeInfo{
					name:     volumeName,
					hostPath: hostPath,
					isPath:   isPath,
				}
			}
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

// parseVolume parses Docker Compose volume format and detects if it's a path or named volume
// Returns: mountPath, volumeName, hostPath (or original name), isPath
func parseVolume(vol string) (mountPath, volumeName, hostPath string, isPath bool) {
	// Handle Windows paths specially to avoid splitting on drive letter colon
	var source string

	// Check for Windows absolute path (C:/, D:/, etc.)
	if len(vol) >= 3 && vol[1] == ':' && (vol[2] == '/' || vol[2] == '\\') {
		// Find the container path after the Windows path
		// Format: C:/path:/container/path
		idx := strings.Index(vol[3:], ":")
		if idx != -1 {
			source = vol[:3+idx]
			mountPath = vol[3+idx+1:]
		} else {
			// No container path, treat whole thing as source
			return vol, "unnamed-vol", "", false
		}
	} else {
		// Normal split for non-Windows paths
		parts := strings.Split(vol, ":")
		if len(parts) >= 2 {
			source = parts[0]
			mountPath = parts[1]
		} else {
			// Anonymous volume
			return vol, "unnamed-vol", "", false
		}
	}

	// Check if source is a path (absolute or relative)
	isPath = isHostPath(source)

	if isPath {
		// It's a host path - use it directly and create a safe volume name
		hostPath = source
		// Convert path to safe volume name
		volumeName = pathToVolumeName(source)
	} else {
		// It's a named volume
		hostPath = source
		volumeName = source
		isPath = false
	}

	return mountPath, volumeName, hostPath, isPath
}

// isHostPath determines if a string is a file system path (absolute or relative)
func isHostPath(path string) bool {
	// Absolute path check (Unix-style)
	if strings.HasPrefix(path, "/") {
		return true
	}

	// Relative path check
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") {
		return true
	}

	// Windows absolute path check (C:\ or similar)
	if len(path) >= 2 && path[1] == ':' {
		return true
	}

	// Check if it contains path separators (might be relative without ./)
	if strings.Contains(path, "/") || strings.Contains(path, "\\") {
		return true
	}

	return false
}

// pathToVolumeName converts a file path to a valid Kubernetes volume name
func pathToVolumeName(path string) string {
	// Clean the path
	cleaned := filepath.Clean(path)

	// Replace separators and special characters
	volumeName := strings.ReplaceAll(cleaned, ":", "-") // Windows drive letters
	volumeName = strings.ReplaceAll(volumeName, "/", "-")
	volumeName = strings.ReplaceAll(volumeName, "\\", "-")
	volumeName = strings.ReplaceAll(volumeName, ".", "-")
	volumeName = strings.ReplaceAll(volumeName, "_", "-")
	volumeName = strings.ToLower(volumeName)

	// Remove leading/trailing dashes
	volumeName = strings.Trim(volumeName, "-")

	// Ensure it doesn't start with a number or special char
	if len(volumeName) > 0 && (volumeName[0] == '-' || (volumeName[0] >= '0' && volumeName[0] <= '9')) {
		volumeName = "vol-" + volumeName
	}

	// Limit length (Kubernetes names can't exceed 253 chars)
	if len(volumeName) > 63 {
		volumeName = volumeName[:63]
		volumeName = strings.TrimRight(volumeName, "-")
	}

	// Fallback if somehow empty
	if volumeName == "" {
		volumeName = "volume"
	}

	return volumeName
}

// parseUser parses user:group format
func parseUser(user string) (uid, gid string) {
	parts := strings.Split(user, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return user, ""
}
