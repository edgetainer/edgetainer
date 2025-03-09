package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

// KeyPair represents an SSH key pair
type KeyPair struct {
	PrivateKey     string // PEM encoded private key
	PublicKey      string // OpenSSH format public key
	AuthorizedKey  string // Line for authorized_keys file
	PrivateKeyPath string // Path to the private key file (if saved)
	PublicKeyPath  string // Path to the public key file (if saved)
}

// GenerateKeyPair creates a new SSH key pair
func GenerateKeyPair(deviceID string, bits int) (*KeyPair, error) {
	if bits == 0 {
		bits = 4096 // Default to 4096 bits
	}

	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Convert private key to PEM format
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	// Convert to SSH public key
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to public key: %w", err)
	}

	// Get public key in OpenSSH authorized_keys format
	pubKeyStr := string(ssh.MarshalAuthorizedKey(publicKey))
	pubKeyStr = pubKeyStr[:len(pubKeyStr)-1] // Remove trailing newline

	// Create the authorized_keys line with the device ID as a prefix
	authKeyStr := fmt.Sprintf("%s %s", deviceID, pubKeyStr)

	return &KeyPair{
		PrivateKey:    string(privateKeyPEM),
		PublicKey:     pubKeyStr,
		AuthorizedKey: authKeyStr,
	}, nil
}

// SaveKeyPair saves the key pair to files
func SaveKeyPair(kp *KeyPair, baseDir, keyName string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Set file paths
	privateKeyPath := filepath.Join(baseDir, keyName)
	publicKeyPath := filepath.Join(baseDir, keyName+".pub")

	// Save private key
	if err := os.WriteFile(privateKeyPath, []byte(kp.PrivateKey), 0600); err != nil {
		return fmt.Errorf("failed to save private key: %w", err)
	}

	// Save public key
	if err := os.WriteFile(publicKeyPath, []byte(kp.PublicKey), 0644); err != nil {
		return fmt.Errorf("failed to save public key: %w", err)
	}

	// Update paths in the key pair
	kp.PrivateKeyPath = privateKeyPath
	kp.PublicKeyPath = publicKeyPath

	return nil
}

// AddToAuthorizedKeys adds the public key to the authorized_keys file
func AddToAuthorizedKeys(kp *KeyPair, authorizedKeysDir, deviceID string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(authorizedKeysDir, 0755); err != nil {
		return fmt.Errorf("failed to create authorized_keys directory: %w", err)
	}

	// Create the device-specific authorized_keys entry
	deviceKeyPath := filepath.Join(authorizedKeysDir, deviceID)
	if err := os.WriteFile(deviceKeyPath, []byte(kp.AuthorizedKey), 0644); err != nil {
		return fmt.Errorf("failed to write device authorized_keys entry: %w", err)
	}

	// Regenerate the main authorized_keys file from all entries
	entries, err := os.ReadDir(authorizedKeysDir)
	if err != nil {
		return fmt.Errorf("failed to read authorized_keys directory: %w", err)
	}

	var allKeysContent string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := os.ReadFile(filepath.Join(authorizedKeysDir, entry.Name()))
		if err != nil {
			continue // Skip files we can't read
		}
		allKeysContent += string(content) + "\n"
	}

	// Write the combined file
	authorizedKeysPath := filepath.Join(filepath.Dir(authorizedKeysDir), "authorized_keys")
	if err := os.WriteFile(authorizedKeysPath, []byte(allKeysContent), 0644); err != nil {
		return fmt.Errorf("failed to write authorized_keys file: %w", err)
	}

	return nil
}
