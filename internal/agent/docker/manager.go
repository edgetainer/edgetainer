package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/edgetainer/edgetainer/internal/shared/logging"
)

// ContainerState represents the state of a container
type ContainerState string

const (
	// ContainerRunning indicates the container is running
	ContainerRunning ContainerState = "running"
	// ContainerStopped indicates the container is stopped
	ContainerStopped ContainerState = "stopped"
	// ContainerRestarting indicates the container is restarting
	ContainerRestarting ContainerState = "restarting"
	// ContainerCreated indicates the container is created but not started
	ContainerCreated ContainerState = "created"
	// ContainerExited indicates the container has exited
	ContainerExited ContainerState = "exited"
	// ContainerUnknown indicates the container state is unknown
	ContainerUnknown ContainerState = "unknown"
)

// Container represents a Docker container
type Container struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Image      string            `json:"image"`
	State      ContainerState    `json:"state"`
	Status     string            `json:"status"`
	Created    string            `json:"created"`
	Ports      map[string]string `json:"ports"`
	VolumesRaw []string          `json:"volumes_raw"`
}

// Application represents a Docker Compose application
type Application struct {
	Name       string            `json:"name"`
	Path       string            `json:"path"`
	Containers []Container       `json:"containers"`
	EnvVars    map[string]string `json:"env_vars"`
	Version    string            `json:"version"`
}

// Manager handles Docker operations
type Manager struct {
	ctx          context.Context
	cancelFunc   context.CancelFunc
	composeDir   string
	networkName  string
	logger       *logging.Logger
	mu           sync.Mutex
	applications map[string]*Application
}

// NewManager creates a new Docker manager
func NewManager(ctx context.Context, composeDir, networkName string) (*Manager, error) {
	managerCtx, cancel := context.WithCancel(ctx)

	// Ensure the compose directory exists
	if err := os.MkdirAll(composeDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create compose directory: %w", err)
	}

	return &Manager{
		ctx:          managerCtx,
		cancelFunc:   cancel,
		composeDir:   composeDir,
		networkName:  networkName,
		logger:       logging.WithComponent("docker-manager"),
		applications: make(map[string]*Application),
	}, nil
}

// Start initializes the Docker manager
func (m *Manager) Start() error {
	m.logger.Info("Docker manager starting")

	// Ensure Docker is running
	if err := m.checkDockerAvailability(); err != nil {
		return fmt.Errorf("docker is not available: %w", err)
	}

	// Create the Docker network if it doesn't exist
	if err := m.ensureNetworkExists(); err != nil {
		return fmt.Errorf("failed to create Docker network: %w", err)
	}

	// Load existing applications
	if err := m.loadExistingApplications(); err != nil {
		m.logger.Error(fmt.Sprintf("Failed to load existing applications: %v", err), err)
		// Continue anyway, non-fatal
	}

	return nil
}

// Stop gracefully shuts down the Docker manager
func (m *Manager) Stop() {
	m.logger.Info("Docker manager stopping")
	m.cancelFunc()
}

// DeployApplication deploys a Docker Compose application
func (m *Manager) DeployApplication(name, composeYAML, version string, envVars map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	appDir := filepath.Join(m.composeDir, name)

	// Create application directory if it doesn't exist
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("failed to create application directory: %w", err)
	}

	// Create docker-compose.yml file
	composeFile := filepath.Join(appDir, "docker-compose.yml")
	if err := os.WriteFile(composeFile, []byte(composeYAML), 0644); err != nil {
		return fmt.Errorf("failed to write docker-compose.yml: %w", err)
	}

	// Create .env file with environment variables
	if len(envVars) > 0 {
		envContent := ""
		for key, value := range envVars {
			envContent += fmt.Sprintf("%s=%s\n", key, value)
		}

		envFile := filepath.Join(appDir, ".env")
		if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
			return fmt.Errorf("failed to write .env file: %w", err)
		}
	}

	// Pull images
	m.logger.Info(fmt.Sprintf("Pulling images for application %s", name))
	cmd := exec.Command("docker-compose", "-f", composeFile, "pull")
	cmd.Dir = appDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to pull images: %v - %s", err, string(output))
	}

	// Start application
	m.logger.Info(fmt.Sprintf("Starting application %s", name))
	cmd = exec.Command("docker-compose", "-f", composeFile, "up", "-d")
	cmd.Dir = appDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start application: %v - %s", err, string(output))
	}

	// Get containers
	containers, err := m.getContainers(name, appDir)
	if err != nil {
		m.logger.Error(fmt.Sprintf("Failed to get containers for application %s: %v", name, err), err)
		// Continue anyway, non-fatal
	}

	// Register application
	m.applications[name] = &Application{
		Name:       name,
		Path:       appDir,
		Containers: containers,
		EnvVars:    envVars,
		Version:    version,
	}

	m.logger.Info(fmt.Sprintf("Successfully deployed application %s version %s", name, version))
	return nil
}

// RemoveApplication removes a Docker Compose application
func (m *Manager) RemoveApplication(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	app, exists := m.applications[name]
	if !exists {
		return fmt.Errorf("application %s not found", name)
	}

	// Stop and remove containers
	m.logger.Info(fmt.Sprintf("Stopping application %s", name))
	cmd := exec.Command("docker-compose", "-f", filepath.Join(app.Path, "docker-compose.yml"), "down", "--remove-orphans")
	cmd.Dir = app.Path
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop application: %v - %s", err, string(output))
	}

	// Remove application directory
	if err := os.RemoveAll(app.Path); err != nil {
		m.logger.Warn(fmt.Sprintf("Failed to remove application directory %s: %v", app.Path, err))
		// Continue anyway, non-fatal
	}

	// Unregister application
	delete(m.applications, name)

	m.logger.Info(fmt.Sprintf("Successfully removed application %s", name))
	return nil
}

// RestartContainer restarts a specific container
func (m *Manager) RestartContainer(appName, containerName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	app, exists := m.applications[appName]
	if !exists {
		return fmt.Errorf("application %s not found", appName)
	}

	// Find the container
	found := false
	for _, container := range app.Containers {
		if container.Name == containerName {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("container %s not found in application %s", containerName, appName)
	}

	// Restart the container
	m.logger.Info(fmt.Sprintf("Restarting container %s in application %s", containerName, appName))
	cmd := exec.Command("docker-compose", "-f", filepath.Join(app.Path, "docker-compose.yml"), "restart", containerName)
	cmd.Dir = app.Path
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to restart container: %v - %s", err, string(output))
	}

	m.logger.Info(fmt.Sprintf("Successfully restarted container %s in application %s", containerName, appName))
	return nil
}

// GetApplications returns all registered applications
func (m *Manager) GetApplications() map[string]*Application {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Return a copy to avoid race conditions
	apps := make(map[string]*Application)
	for name, app := range m.applications {
		appCopy := *app
		containersCopy := make([]Container, len(app.Containers))
		copy(containersCopy, app.Containers)
		appCopy.Containers = containersCopy

		envVarsCopy := make(map[string]string)
		for k, v := range app.EnvVars {
			envVarsCopy[k] = v
		}
		appCopy.EnvVars = envVarsCopy

		apps[name] = &appCopy
	}

	return apps
}

// UpdateEnvironmentVariables updates environment variables for an application
func (m *Manager) UpdateEnvironmentVariables(appName string, envVars map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	app, exists := m.applications[appName]
	if !exists {
		return fmt.Errorf("application %s not found", appName)
	}

	// Update .env file
	envContent := ""
	for key, value := range envVars {
		envContent += fmt.Sprintf("%s=%s\n", key, value)
	}

	envFile := filepath.Join(app.Path, ".env")
	if err := os.WriteFile(envFile, []byte(envContent), 0644); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	// Update application
	app.EnvVars = envVars

	m.logger.Info(fmt.Sprintf("Successfully updated environment variables for application %s", appName))
	return nil
}

// GetContainerLogs returns logs for a specific container
func (m *Manager) GetContainerLogs(appName, containerName string, lines int) (string, error) {
	m.mu.Lock()
	app, exists := m.applications[appName]
	m.mu.Unlock()

	if !exists {
		return "", fmt.Errorf("application %s not found", appName)
	}

	// Get container logs
	args := []string{
		"-f", filepath.Join(app.Path, "docker-compose.yml"),
		"logs",
		"--tail", fmt.Sprintf("%d", lines),
		containerName,
	}

	cmd := exec.Command("docker-compose", args...)
	cmd.Dir = app.Path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}

	return string(output), nil
}

// checkDockerAvailability checks if Docker is available
func (m *Manager) checkDockerAvailability() error {
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker is not installed or not running: %v - %s", err, string(output))
	}

	m.logger.Info(fmt.Sprintf("Docker version: %s", strings.TrimSpace(string(output))))

	cmd = exec.Command("docker-compose", "version", "--short")
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker-compose is not installed: %v - %s", err, string(output))
	}

	m.logger.Info(fmt.Sprintf("Docker Compose version: %s", strings.TrimSpace(string(output))))

	return nil
}

// ensureNetworkExists creates the Docker network if it doesn't exist
func (m *Manager) ensureNetworkExists() error {
	cmd := exec.Command("docker", "network", "inspect", m.networkName)
	if err := cmd.Run(); err == nil {
		// Network already exists
		return nil
	}

	// Create the network
	cmd = exec.Command("docker", "network", "create", m.networkName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create Docker network: %v - %s", err, string(output))
	}

	m.logger.Info(fmt.Sprintf("Created Docker network: %s", m.networkName))
	return nil
}

// loadExistingApplications loads existing Docker Compose applications
func (m *Manager) loadExistingApplications() error {
	// Read compose directory
	files, err := ioutil.ReadDir(m.composeDir)
	if err != nil {
		return fmt.Errorf("failed to read compose directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		appName := file.Name()
		appDir := filepath.Join(m.composeDir, appName)

		// Check if docker-compose.yml exists
		composeFile := filepath.Join(appDir, "docker-compose.yml")
		if _, err := os.Stat(composeFile); os.IsNotExist(err) {
			continue
		}

		// Load environment variables
		envVars := make(map[string]string)
		envFile := filepath.Join(appDir, ".env")
		if _, err := os.Stat(envFile); err == nil {
			// Parse .env file
			envData, err := ioutil.ReadFile(envFile)
			if err == nil {
				lines := strings.Split(string(envData), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "" || strings.HasPrefix(line, "#") {
						continue
					}

					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						envVars[parts[0]] = parts[1]
					}
				}
			}
		}

		// Get containers
		containers, err := m.getContainers(appName, appDir)
		if err != nil {
			m.logger.Error(fmt.Sprintf("Failed to get containers for application %s: %v", appName, err), err)
			// Continue anyway, non-fatal
			containers = []Container{}
		}

		// Register application
		m.applications[appName] = &Application{
			Name:       appName,
			Path:       appDir,
			Containers: containers,
			EnvVars:    envVars,
			Version:    "unknown", // Cannot determine version without metadata
		}

		m.logger.Info(fmt.Sprintf("Loaded existing application %s with %d containers", appName, len(containers)))
	}

	return nil
}

// getContainers gets containers for an application
func (m *Manager) getContainers(appName, appDir string) ([]Container, error) {
	cmd := exec.Command("docker-compose", "-f", filepath.Join(appDir, "docker-compose.yml"), "ps", "--format", "json")
	cmd.Dir = appDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get containers: %v - %s", err, string(output))
	}

	// Parse output
	var result []map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		// Fallback for older versions of docker-compose that don't support JSON output
		return m.getContainersLegacy(appName, appDir)
	}

	// Convert to Container structs
	containers := make([]Container, 0, len(result))
	for _, item := range result {
		container := Container{
			Name:       fmt.Sprintf("%v", item["Name"]),
			Image:      fmt.Sprintf("%v", item["Image"]),
			State:      ContainerState(fmt.Sprintf("%v", item["State"])),
			Status:     fmt.Sprintf("%v", item["Status"]),
			Ports:      make(map[string]string),
			VolumesRaw: make([]string, 0),
		}

		containers = append(containers, container)
	}

	return containers, nil
}

// getContainersLegacy gets containers for an application using legacy format
func (m *Manager) getContainersLegacy(appName, appDir string) ([]Container, error) {
	// This is a simplified implementation for older docker-compose versions
	// In a real implementation, you would parse the output of docker-compose ps
	cmd := exec.Command("docker-compose", "-f", filepath.Join(appDir, "docker-compose.yml"), "ps", "-q")
	cmd.Dir = appDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get container IDs: %v - %s", err, string(output))
	}

	containerIDs := strings.Split(strings.TrimSpace(string(output)), "\n")
	containers := make([]Container, 0, len(containerIDs))

	for _, id := range containerIDs {
		if id == "" {
			continue
		}

		cmd := exec.Command("docker", "inspect", id)
		output, err := cmd.CombinedOutput()
		if err != nil {
			m.logger.Error(fmt.Sprintf("Failed to inspect container %s: %v", id, err), err)
			continue
		}

		var inspectResult []map[string]interface{}
		if err := json.Unmarshal(output, &inspectResult); err != nil {
			m.logger.Error(fmt.Sprintf("Failed to parse inspect output for container %s: %v", id, err), err)
			continue
		}

		if len(inspectResult) == 0 {
			continue
		}

		// Extract container information
		info := inspectResult[0]
		name := fmt.Sprintf("%v", info["Name"])
		if strings.HasPrefix(name, "/") {
			name = name[1:] // Remove leading slash
		}

		state := ContainerUnknown
		if stateInfo, ok := info["State"].(map[string]interface{}); ok {
			if running, ok := stateInfo["Running"].(bool); ok && running {
				state = ContainerRunning
			} else if status, ok := stateInfo["Status"].(string); ok {
				switch status {
				case "created":
					state = ContainerCreated
				case "exited":
					state = ContainerExited
				case "restarting":
					state = ContainerRestarting
				}
			}
		}

		image := ""
		if config, ok := info["Config"].(map[string]interface{}); ok {
			if img, ok := config["Image"].(string); ok {
				image = img
			}
		}

		container := Container{
			ID:         id,
			Name:       name,
			Image:      image,
			State:      state,
			Status:     fmt.Sprintf("%v", state),
			Ports:      make(map[string]string),
			VolumesRaw: make([]string, 0),
		}

		containers = append(containers, container)
	}

	return containers, nil
}
