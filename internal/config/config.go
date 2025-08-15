package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// InitInteractive creates or updates configuration interactively
func InitInteractive(path string) error {
	if path == "" {
		path = DefaultPath()
	}

	reader := bufio.NewReader(os.Stdin)
	
	// Check if config exists
	var cfg *Config
	if _, err := os.Stat(path); err == nil {
		cfg, err = loadFromFile(path)
		if err != nil {
			fmt.Printf("Warning: existing config is invalid, creating new one: %v\n", err)
			cfg = &Config{
				Accounts: make(map[string]AccountConfig),
				Orgs: make(map[string]string),
				Folders: make(map[string][]string),
			}
		} else {
			fmt.Println("Updating existing configuration...")
		}
	} else {
		fmt.Println("Creating new configuration...")
		cfg = &Config{
			Accounts: make(map[string]AccountConfig),
			Orgs: make(map[string]string),
			Folders: make(map[string][]string),
		}
	}

	// Get base directory
	fmt.Printf("Base directory for repositories [%s]: ", defaultBaseDir())
	baseDir, _ := reader.ReadString('\n')
	baseDir = strings.TrimSpace(baseDir)
	if baseDir == "" {
		baseDir = defaultBaseDir()
	}
	cfg.BaseDir = baseDir

	// Add/update accounts
	fmt.Println("\nGitHub Account Configuration")
	fmt.Println("(Press Enter with no username to finish)")
	
	for {
		fmt.Print("\nGitHub username: ")
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)
		if username == "" {
			break
		}

		// Get account type
		fmt.Print("Account type (personal/work/school/org) [personal]: ")
		accountType, _ := reader.ReadString('\n')
		accountType = strings.TrimSpace(accountType)
		if accountType == "" {
			accountType = "personal"
		}

		// Get SSH host
		fmt.Printf("SSH host config name [github-%s]: ", username)
		sshHost, _ := reader.ReadString('\n')
		sshHost = strings.TrimSpace(sshHost)
		if sshHost == "" {
			sshHost = fmt.Sprintf("github-%s", username)
		}

		// Get email
		fmt.Printf("Git email for this account: ")
		email, _ := reader.ReadString('\n')
		email = strings.TrimSpace(email)

		cfg.Accounts[username] = AccountConfig{
			Type:    accountType,
			SSHHost: sshHost,
			Email:   email,
		}

		fmt.Printf("✓ Added account: %s\n", username)
	}

	// Save configuration
	if err := cfg.Save(path); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("\n✓ Configuration saved to: %s\n", path)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Ensure your SSH config has entries for each SSH host")
	fmt.Println("2. Run 'ds scan' to discover repositories")
	fmt.Println("3. Run 'ds status' to see repository status")
	
	return nil
}