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

func TestVolumePathDetection(t *testing.T) {
	tests := []struct {
		name           string
		volumeSpec     string
		expectIsPath   bool
		expectHostPath string
		expectVolName  string
	}{
		{
			name:           "Absolute Unix path",
			volumeSpec:     "/host/data:/container/data",
			expectIsPath:   true,
			expectHostPath: "/host/data",
			expectVolName:  "host-data",
		},
		{
			name:           "Relative path with ./",
			volumeSpec:     "./config:/etc/config",
			expectIsPath:   true,
			expectHostPath: "./config",
			expectVolName:  "config",
		},
		{
			name:           "Relative path with ../",
			volumeSpec:     "../data:/data",
			expectIsPath:   true,
			expectHostPath: "../data",
			expectVolName:  "data",
		},
		{
			name:           "Named volume",
			volumeSpec:     "db-data:/var/lib/mysql",
			expectIsPath:   false,
			expectHostPath: "db-data",
			expectVolName:  "db-data",
		},
		{
			name:           "Path with subdirectories",
			volumeSpec:     "./app/config/settings:/etc/app",
			expectIsPath:   true,
			expectHostPath: "./app/config/settings",
			expectVolName:  "app-config-settings",
		},
		{
			name:           "Windows absolute path",
			volumeSpec:     "C:/data:/data",
			expectIsPath:   true,
			expectHostPath: "C:/data",
			expectVolName:  "c--data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mountPath, volName, hostPath, isPath := parseVolume(tt.volumeSpec)

			if isPath != tt.expectIsPath {
				t.Errorf("isPath = %v, want %v", isPath, tt.expectIsPath)
			}

			if hostPath != tt.expectHostPath {
				t.Errorf("hostPath = %s, want %s", hostPath, tt.expectHostPath)
			}

			if volName != tt.expectVolName {
				t.Errorf("volumeName = %s, want %s", volName, tt.expectVolName)
			}

			// Verify mount path is extracted correctly
			parts := strings.Split(tt.volumeSpec, ":")
			if len(parts) >= 2 && mountPath != parts[1] {
				t.Errorf("mountPath = %s, want %s", mountPath, parts[1])
			}
		})
	}
}

func TestIsHostPath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/absolute/path", true},
		{"./relative/path", true},
		{"../parent/path", true},
		{"C:/windows/path", true},
		{"D:\\windows\\path", true},
		{"named-volume", false},
		{"simple", false},
		{"/", true},
		{"./", true},
		{"sub/dir/path", true}, // contains slash
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isHostPath(tt.path)
			if result != tt.expected {
				t.Errorf("isHostPath(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestPathToVolumeName(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/host/data", "host-data"},
		{"./config", "config"},
		{"../data", "data"},
		{"/var/lib/mysql", "var-lib-mysql"},
		{"./app/config/file.txt", "app-config-file-txt"},
		{"C:/data", "c--data"},
		{"123-start-with-number", "vol-123-start-with-number"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := pathToVolumeName(tt.path)
			if result != tt.expected {
				t.Errorf("pathToVolumeName(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestGenerateWithHostPathVolumes(t *testing.T) {
	compose := &types.ComposeFile{
		Version: "3",
		Services: map[string]types.Service{
			"web": {
				Image: "nginx:latest",
				Volumes: []string{
					"/host/data:/usr/share/nginx/html",
					"./config:/etc/nginx",
					"web-data:/var/www",
				},
			},
		},
		Volumes: map[string]types.Volume{
			"web-data": {},
		},
	}

	gen := NewGenerator(compose, "test-pod")
	yaml, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify hostPath volumes
	if !strings.Contains(yaml, "hostPath:") {
		t.Error("Expected hostPath volume definition")
	}

	if !strings.Contains(yaml, "path: /host/data") {
		t.Error("Expected absolute path /host/data")
	}

	if !strings.Contains(yaml, "path: ./config") {
		t.Error("Expected relative path ./config")
	}

	// Verify PVC for named volume
	if !strings.Contains(yaml, "persistentVolumeClaim:") {
		t.Error("Expected PVC for named volume")
	}

	if !strings.Contains(yaml, "claimName: web-data") {
		t.Error("Expected PVC claim name web-data")
	}

	// Verify all volume names are present
	if !strings.Contains(yaml, "name: host-data") {
		t.Error("Expected volume name host-data")
	}

	if !strings.Contains(yaml, "name: config") {
		t.Error("Expected volume name config")
	}

	if !strings.Contains(yaml, "name: web-data") {
		t.Error("Expected volume name web-data")
	}
}
