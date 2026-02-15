// Package types defines data structures for Docker Compose files and provides
// helper methods for handling flexible field types (maps vs arrays).
package types

// ComposeFile represents a Docker Compose file structure
type ComposeFile struct {
	Version  string             `yaml:"version"`
	Services map[string]Service `yaml:"services"`
	Networks map[string]Network `yaml:"networks,omitempty"`
	Volumes  map[string]Volume  `yaml:"volumes,omitempty"`
}

// Service represents a service definition in Docker Compose
type Service struct {
	Image         string            `yaml:"image,omitempty"`
	ContainerName string            `yaml:"container_name,omitempty"`
	Restart       string            `yaml:"restart,omitempty"`
	WorkingDir    string            `yaml:"working_dir,omitempty"`
	User          string            `yaml:"user,omitempty"`
	Hostname      string            `yaml:"hostname,omitempty"`
	Privileged    bool              `yaml:"privileged,omitempty"`
	Build         interface{}       `yaml:"build,omitempty"`
	Ports         []string          `yaml:"ports,omitempty"`
	Environment   interface{}       `yaml:"environment,omitempty"`
	Volumes       []string          `yaml:"volumes,omitempty"`
	Networks      interface{}       `yaml:"networks,omitempty"`
	DependsOn     interface{}       `yaml:"depends_on,omitempty"`
	Command       interface{}       `yaml:"command,omitempty"`
	Entrypoint    interface{}       `yaml:"entrypoint,omitempty"`
	Labels        map[string]string `yaml:"labels,omitempty"`
	CapAdd        []string          `yaml:"cap_add,omitempty"`
	CapDrop       []string          `yaml:"cap_drop,omitempty"`
}

// Network represents a network definition
type Network struct {
	Driver   string            `yaml:"driver,omitempty"`
	External bool              `yaml:"external,omitempty"`
	Labels   map[string]string `yaml:"labels,omitempty"`
}

// Volume represents a volume definition
type Volume struct {
	Driver   string            `yaml:"driver,omitempty"`
	External bool              `yaml:"external,omitempty"`
	Labels   map[string]string `yaml:"labels,omitempty"`
}

// EnvironmentMap converts environment interface to map
func (s *Service) EnvironmentMap() map[string]string {
	env := make(map[string]string)

	switch v := s.Environment.(type) {
	case map[string]interface{}:
		for key, val := range v {
			if str, ok := val.(string); ok {
				env[key] = str
			}
		}
	case map[interface{}]interface{}:
		for key, val := range v {
			if keyStr, ok := key.(string); ok {
				if valStr, ok := val.(string); ok {
					env[keyStr] = valStr
				}
			}
		}
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				// Handle KEY=VALUE format
				if idx := findEquals(str); idx > 0 {
					env[str[:idx]] = str[idx+1:]
				}
			}
		}
	}

	return env
}

// NetworksList returns networks as a list of strings
func (s *Service) NetworksList() []string {
	var networks []string

	switch v := s.Networks.(type) {
	case []interface{}:
		for _, n := range v {
			if str, ok := n.(string); ok {
				networks = append(networks, str)
			}
		}
	case map[string]interface{}:
		for name := range v {
			networks = append(networks, name)
		}
	case map[interface{}]interface{}:
		for name := range v {
			if str, ok := name.(string); ok {
				networks = append(networks, str)
			}
		}
	}

	return networks
}

// DependsOnList returns dependencies as a list of strings
func (s *Service) DependsOnList() []string {
	var deps []string

	switch v := s.DependsOn.(type) {
	case []interface{}:
		for _, d := range v {
			if str, ok := d.(string); ok {
				deps = append(deps, str)
			}
		}
	case map[string]interface{}:
		for name := range v {
			deps = append(deps, name)
		}
	case map[interface{}]interface{}:
		for name := range v {
			if str, ok := name.(string); ok {
				deps = append(deps, str)
			}
		}
	}

	return deps
}

// CommandList returns command as a list of strings
func (s *Service) CommandList() []string {
	switch v := s.Command.(type) {
	case string:
		return []string{v}
	case []interface{}:
		var cmd []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				cmd = append(cmd, str)
			}
		}
		return cmd
	}
	return nil
}

// EntrypointList returns entrypoint as a list of strings
func (s *Service) EntrypointList() []string {
	switch v := s.Entrypoint.(type) {
	case string:
		return []string{v}
	case []interface{}:
		var ep []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				ep = append(ep, str)
			}
		}
		return ep
	}
	return nil
}

func findEquals(s string) int {
	for i, c := range s {
		if c == '=' {
			return i
		}
	}
	return -1
}
