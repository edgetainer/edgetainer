package provisioning

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

// TemplateData contains variables to be used in Butane templates
type TemplateData struct {
	DeviceID      string
	SSHPrivateKey string
	ServerHost    string
	ServerPort    int
	SSHPort       int
	// Add more fields as needed for templating
}

// RenderButaneTemplate takes a template path and data, and returns the rendered Butane config
func RenderButaneTemplate(templatePath string, data *TemplateData) (string, error) {
	// Read the template file
	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse the template
	tmpl, err := template.New(filepath.Base(templatePath)).Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ConvertButaneToIgnition takes Butane YAML and converts it to Ignition JSON
// This function requires the butane CLI tool to be installed
func ConvertButaneToIgnition(butaneConfig string) (string, error) {
	// Create a temporary file for the Butane config
	tempFile, err := os.CreateTemp("", "butane-*.bu")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	// Write the Butane config to the temporary file
	if _, err := tempFile.WriteString(butaneConfig); err != nil {
		return "", fmt.Errorf("failed to write to temporary file: %w", err)
	}
	tempFile.Close()

	// Execute the butane CLI tool
	cmd := exec.Command("butane", "--pretty", "--strict", tempFile.Name())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("butane conversion failed: %s\nError: %w", stderr.String(), err)
	}

	return stdout.String(), nil
}

// GenerateIgnitionConfig generates the final Ignition JSON from the template and data
func GenerateIgnitionConfig(templatePath string, data *TemplateData) (string, error) {
	// Render the Butane template
	butaneConfig, err := RenderButaneTemplate(templatePath, data)
	if err != nil {
		return "", fmt.Errorf("failed to render butane template: %w", err)
	}

	// Convert to Ignition JSON
	ignitionJSON, err := ConvertButaneToIgnition(butaneConfig)
	if err != nil {
		return "", fmt.Errorf("failed to convert to ignition JSON: %w", err)
	}

	return ignitionJSON, nil
}
