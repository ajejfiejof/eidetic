package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config defines user settings and index directories.
type Config struct {
	DataDir      string   `json:"data_dir"`
	WatchDirs    []string `json:"watch_dirs"`
	IgnoreRules  []string `json:"ignore_rules"`
	MaxDocSizeKB int      `json:"max_doc_size_kb"`
	MaxResults   int      `json:"max_results"`
}

// DefaultConfig initializes sane defaults adhering to XDG standards.
func DefaultConfig() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	dataDir := filepath.Join(home, ".local", "share", "eidetic")

	return &Config{
		DataDir: dataDir,
		WatchDirs: []string{
			filepath.Join(home, "Notes"),
			filepath.Join(home, ".config"),
		},
		IgnoreRules: []string{
			"node_modules",
			".git",
			"__pycache__",
			".cache",
			"target",
			"dist",
			".venv",
		},
		MaxDocSizeKB: 512,
		MaxResults:   50,
	}
}

// ConfigPath returns the standard configuration file path.
func ConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "eidetic", "config.json")
}

// LoadConfig loads the configuration or writes defaults if missing.
func LoadConfig() (*Config, error) {
	path := ConfigPath()
	cfg := DefaultConfig()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		_ = os.MkdirAll(filepath.Dir(path), 0755)
		_ = os.MkdirAll(cfg.DataDir, 0755)
		data, _ := json.MarshalIndent(cfg, "", "  ")
		_ = os.WriteFile(path, data, 0644)
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return DefaultConfig(), err
	}

	_ = os.MkdirAll(cfg.DataDir, 0755)
	return cfg, nil
}
