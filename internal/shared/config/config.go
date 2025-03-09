package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ServerConfig represents the server configuration
type ServerConfig struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
	} `yaml:"database"`
	Auth struct {
		AdminUsername string `yaml:"admin_username"`
		AdminPassword string `yaml:"admin_password"`
		AdminEmail    string `yaml:"admin_email"`
	} `yaml:"auth"`
	SSH struct {
		Port        int    `yaml:"port"`
		HostKeyPath string `yaml:"host_key_path"`
		StartPort   int    `yaml:"start_port"`
		EndPort     int    `yaml:"end_port"`
	} `yaml:"ssh"`
	Logging struct {
		Level   string `yaml:"level"`
		LogFile string `yaml:"log_file"`
	} `yaml:"logging"`
}

// AgentConfig represents the agent configuration
type AgentConfig struct {
	Device struct {
		ID   string `yaml:"id"`
		Name string `yaml:"name"`
	} `yaml:"device"`
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	SSH struct {
		Port int    `yaml:"port"`
		Key  string `yaml:"key"`
	} `yaml:"ssh"`
	Docker struct {
		ComposeDir  string `yaml:"compose_dir"`
		NetworkName string `yaml:"network_name"`
	} `yaml:"docker"`
	Logging struct {
		Level   string `yaml:"level"`
		LogFile string `yaml:"log_file"`
	} `yaml:"logging"`
}

// LoadServerConfig loads the server configuration from a file
func LoadServerConfig(path string) (*ServerConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg ServerConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults for missing values
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.SSH.Port == 0 {
		cfg.SSH.Port = 2222
	}
	if cfg.SSH.HostKeyPath == "" {
		cfg.SSH.HostKeyPath = "ssh_host_key"
	}
	if cfg.SSH.StartPort == 0 {
		cfg.SSH.StartPort = 10000
	}
	if cfg.SSH.EndPort == 0 {
		cfg.SSH.EndPort = 20000
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}

	// Check for environment variables for admin credentials
	if adminUsername := os.Getenv("EDGETAINER_ADMIN_USERNAME"); adminUsername != "" {
		cfg.Auth.AdminUsername = adminUsername
	} else if cfg.Auth.AdminUsername == "" {
		cfg.Auth.AdminUsername = "admin"
	}

	if adminPassword := os.Getenv("EDGETAINER_ADMIN_PASSWORD"); adminPassword != "" {
		cfg.Auth.AdminPassword = adminPassword
	} else if cfg.Auth.AdminPassword == "" {
		cfg.Auth.AdminPassword = "password"
	}

	if adminEmail := os.Getenv("EDGETAINER_ADMIN_EMAIL"); adminEmail != "" {
		cfg.Auth.AdminEmail = adminEmail
	} else if cfg.Auth.AdminEmail == "" {
		cfg.Auth.AdminEmail = "admin@example.com"
	}

	return &cfg, nil
}

// LoadAgentConfig loads the agent configuration from a file
func LoadAgentConfig(path string) (*AgentConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg AgentConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults for missing values
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.SSH.Port == 0 {
		cfg.SSH.Port = 2222
	}
	if cfg.SSH.Key == "" {
		cfg.SSH.Key = "ssh_key"
	}
	if cfg.Docker.ComposeDir == "" {
		cfg.Docker.ComposeDir = "compose"
	}
	if cfg.Docker.NetworkName == "" {
		cfg.Docker.NetworkName = "edgetainer"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}

	return &cfg, nil
}

// CreateDefaultServerConfig creates a default server configuration file
func CreateDefaultServerConfig(path string) error {
	// Create default configuration
	cfg := ServerConfig{}
	cfg.Server.Host = "0.0.0.0"
	cfg.Server.Port = 8080
	cfg.Database.Host = "localhost"
	cfg.Database.Port = 5432
	cfg.Database.User = "postgres"
	cfg.Database.Password = "postgres"
	cfg.Database.DBName = "edgetainer"
	cfg.Auth.AdminUsername = "admin"
	cfg.Auth.AdminPassword = "password"
	cfg.Auth.AdminEmail = "admin@example.com"
	cfg.SSH.Port = 2222
	cfg.SSH.HostKeyPath = "ssh_host_key"
	cfg.SSH.StartPort = 10000
	cfg.SSH.EndPort = 20000
	cfg.Logging.Level = "info"
	cfg.Logging.LogFile = "edgetainer-server.log"

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Marshal and write to file
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// CreateDefaultAgentConfig creates a default agent configuration file
func CreateDefaultAgentConfig(path string) error {
	// Create default configuration
	cfg := AgentConfig{}
	cfg.Device.ID = generateDeviceID()
	cfg.Device.Name = "edgetainer-device"
	cfg.Server.Host = "localhost"
	cfg.Server.Port = 8080
	cfg.SSH.Port = 2222
	cfg.SSH.Key = "ssh_key"
	cfg.Docker.ComposeDir = "compose"
	cfg.Docker.NetworkName = "edgetainer"
	cfg.Logging.Level = "info"
	cfg.Logging.LogFile = "edgetainer-agent.log"

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Marshal and write to file
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SaveServerConfig saves the server configuration to a file
func SaveServerConfig(cfg *ServerConfig, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// SaveAgentConfig saves the agent configuration to a file
func SaveAgentConfig(cfg *AgentConfig, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// generateDeviceID generates a unique device ID
func generateDeviceID() string {
	// Simple implementation - in a real system, use something more robust
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "device"
	}

	// Add a timestamp to ensure uniqueness
	return fmt.Sprintf("%s-%d", hostname, os.Getpid())
}
