// Package config manages the cred configuration file and credential metadata.
//
// Config file location: ~/.config/cred/config.yml
//
// The config stores metadata about registered credentials (key names, namespaces)
// but NEVER stores the actual secret values — those live in the OS keyring.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the top-level configuration.
type Config struct {
	Version    int                  `yaml:"version"`
	Default    string               `yaml:"default_namespace"`
	Namespaces map[string]Namespace `yaml:"namespaces"`
}

// Namespace groups credentials by project or environment.
type Namespace struct {
	Description string            `yaml:"description,omitempty"`
	Keys        map[string]KeyMeta `yaml:"keys"`
}

// KeyMeta stores metadata about a credential (not the secret itself).
type KeyMeta struct {
	Description string `yaml:"description,omitempty"`
	EnvVar      string `yaml:"env_var,omitempty"` // env var name for "run" command
	CreatedAt   string `yaml:"created_at,omitempty"`
	UpdatedAt   string `yaml:"updated_at,omitempty"`
}

// configDir returns the config directory path.
func configDir() (string, error) {
	// Respect XDG_CONFIG_HOME if set
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "cred"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "cred"), nil
}

// ConfigPath returns the full path to the config file.
func ConfigPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yml"), nil
}

// Load reads the config file. Returns a default config if the file doesn't exist.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return defaultConfig(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if cfg.Namespaces == nil {
		cfg.Namespaces = make(map[string]Namespace)
	}

	return &cfg, nil
}

// Save writes the config to disk.
func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(path, data, 0600)
}

// EnsureNamespace creates a namespace if it doesn't exist.
func (c *Config) EnsureNamespace(name string) {
	if _, ok := c.Namespaces[name]; !ok {
		c.Namespaces[name] = Namespace{
			Keys: make(map[string]KeyMeta),
		}
	}
}

// SetKeyMeta registers or updates metadata for a credential key.
func (c *Config) SetKeyMeta(namespace, key string, meta KeyMeta) {
	c.EnsureNamespace(namespace)
	ns := c.Namespaces[namespace]
	if ns.Keys == nil {
		ns.Keys = make(map[string]KeyMeta)
	}
	ns.Keys[key] = meta
	c.Namespaces[namespace] = ns
}

// DeleteKeyMeta removes metadata for a credential key.
func (c *Config) DeleteKeyMeta(namespace, key string) {
	ns, ok := c.Namespaces[namespace]
	if !ok {
		return
	}
	delete(ns.Keys, key)
	c.Namespaces[namespace] = ns
}

// ActiveNamespace returns the namespace to use (flag override or default).
func (c *Config) ActiveNamespace(override string) string {
	if override != "" {
		return override
	}
	if c.Default != "" {
		return c.Default
	}
	return "default"
}

func defaultConfig() *Config {
	return &Config{
		Version:    1,
		Default:    "default",
		Namespaces: make(map[string]Namespace),
	}
}
