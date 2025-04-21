package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config represents the configuration file structure
type Config struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Manager handles configuration file operations
type Manager struct {
	FilePath string
	Config   Config
}

// NewManager creates a new config manager
func NewManager() *Manager {
	configFilePath := os.ExpandEnv("$HOME/.thread-art-cli.json")

	manager := &Manager{
		FilePath: configFilePath,
	}

	// Load config when creating the manager
	manager.Load()

	return manager
}

// Load loads the configuration from disk
func (m *Manager) Load() error {
	file, err := os.Open(m.FilePath)
	if err != nil {
		// If file doesn't exist, that's okay
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("could not open config file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&m.Config); err != nil {
		return fmt.Errorf("could not decode config file: %v", err)
	}

	return nil
}

// Save saves the configuration to disk
func (m *Manager) Save() error {
	file, err := os.Create(m.FilePath)
	if err != nil {
		return fmt.Errorf("could not create config file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(m.Config); err != nil {
		return fmt.Errorf("could not encode config: %v", err)
	}

	return nil
}

// IsTokenValid checks if the current token is valid
func (m *Manager) IsTokenValid() bool {
	if m.Config.AccessToken == "" {
		return false
	}

	return m.Config.ExpiresAt.After(time.Now())
}

// Clear clears the configuration (for logout)
func (m *Manager) Clear() error {
	m.Config = Config{}
	return m.Save()
}
