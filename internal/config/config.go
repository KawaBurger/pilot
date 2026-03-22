package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Defaults DefaultsConfig `yaml:"defaults"`
	Claude  ClaudeConfig  `yaml:"claude"`
}

type ServerConfig struct {
	Port           int      `yaml:"port"`
	Host           string   `yaml:"host"`
	AllowedOrigins []string `yaml:"allowed_origins"`
}

type DefaultsConfig struct {
	CWD string `yaml:"cwd"`
}

type ClaudeConfig struct {
	Home string `yaml:"home"`
}

type Command struct {
	Name     string `yaml:"name" json:"name"`
	Label    string `yaml:"label" json:"label"`
	Template string `yaml:"template" json:"template"`
}

type CommandsConfig struct {
	Commands []Command `yaml:"commands"`
}

// PilotHome returns the path to ~/.pilot
func PilotHome() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pilot")
}

// Load reads config from ~/.pilot/config.yaml, applying defaults for missing values.
func Load() (*Config, error) {
	home, _ := os.UserHomeDir()

	cfg := &Config{
		Server: ServerConfig{
			Port: 8090,
			Host: "127.0.0.1",
		},
		Defaults: DefaultsConfig{
			CWD: home,
		},
		Claude: ClaudeConfig{
			Home: filepath.Join(home, ".claude"),
		},
	}

	path := filepath.Join(PilotHome(), "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadCommands reads commands from ~/.pilot/commands.yaml, applying defaults for missing values.
func LoadCommands() (*CommandsConfig, error) {
	defaults := &CommandsConfig{
		Commands: []Command{
			{Name: "continue", Label: "Continue", Template: "continue"},
			{Name: "commit", Label: "Commit", Template: "commit"},
			{Name: "fix", Label: "Fix", Template: "fix"},
			{Name: "explain", Label: "Explain", Template: "explain"},
			{Name: "test", Label: "Test", Template: "test"},
		},
	}

	path := filepath.Join(PilotHome(), "commands.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaults, nil
		}
		return nil, err
	}

	cfg := &CommandsConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	if len(cfg.Commands) == 0 {
		return defaults, nil
	}

	return cfg, nil
}
