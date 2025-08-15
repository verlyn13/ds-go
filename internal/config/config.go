package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

// Config holds application configuration
type Config struct {
	BaseDir  string                       `yaml:"base_dir" json:"base_dir"`
	Accounts map[string]AccountConfig     `yaml:"accounts" json:"accounts"`
	Orgs     map[string]string           `yaml:"organizations" json:"organizations"`
	Folders  map[string][]string         `yaml:"folder_structure" json:"folder_structure"`
}

// AccountConfig holds account-specific configuration
type AccountConfig struct {
	Type    string `yaml:"type" json:"type"`
	SSHHost string `yaml:"ssh_host" json:"ssh_host"`
	Email   string `yaml:"email" json:"email"`
}

// DefaultPath returns the default config file path using XDG
func DefaultPath() string {
	return filepath.Join(xdg.ConfigHome, "ds", "config.yaml")
}

// Load loads configuration from file or creates default
func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultPath()
	}
	
	// Ensure config directory exists
	configDir := filepath.Dir(path)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("creating config directory: %w", err)
	}
	
	// Try to load existing config
	if _, err := os.Stat(path); err == nil {
		return loadFromFile(path)
	}
	
	// Create default config
	cfg := defaultConfig()
	
	// Save default config
	if err := cfg.Save(path); err != nil {
		return nil, fmt.Errorf("saving default config: %w", err)
	}
	
	return cfg, nil
}

// loadFromFile loads config from YAML or JSON file
func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	
	var cfg Config
	
	// Try YAML first
	if err := yaml.Unmarshal(data, &cfg); err == nil {
		// Set default base dir if not specified
		if cfg.BaseDir == "" {
			cfg.BaseDir = defaultBaseDir()
		}
		return &cfg, nil
	}
	
	// Try JSON
	if err := json.Unmarshal(data, &cfg); err == nil {
		if cfg.BaseDir == "" {
			cfg.BaseDir = defaultBaseDir()
		}
		return &cfg, nil
	}
	
	return nil, fmt.Errorf("invalid config format (not YAML or JSON)")
}

// Save saves configuration to file
func (c *Config) Save(path string) error {
	if path == "" {
		path = DefaultPath()
	}
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	
	// Marshal to YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	
	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}
	
	return nil
}

// defaultBaseDir returns the default base directory
func defaultBaseDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Projects")
}

// defaultConfig returns the default configuration
func defaultConfig() *Config {
	return &Config{
		BaseDir: defaultBaseDir(),
		Accounts: map[string]AccountConfig{
			"verlyn13": {
				Type:    "personal",
				SSHHost: "github-personal",
				Email:   "personal@example.com",
			},
			"jjohnson-47": {
				Type:    "school", 
				SSHHost: "github-work",
				Email:   "school@university.edu",
			},
			"happy-patterns": {
				Type:    "organization",
				SSHHost: "happy-patterns",
				Email:   "happy@patterns.org",
			},
		},
		Orgs: map[string]string{
			"ScopeTechGtHb":        "github-scope",
			"AndroidScopeProjects": "github.com",
			"The-Nash-Group":       "github.com",
			"happy-patterns-org":   "happy-patterns",
		},
		Folders: map[string][]string{
			"personal": {"verlyn13"},
			"school":   {"jjohnson-47"},
			"orgs":     {"happy-patterns", "ScopeTechGtHb", "AndroidScopeProjects"},
		},
	}
}