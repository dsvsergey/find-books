// Package config stores user settings in DataDir()/config.json.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"findbooks/internal/platform"
)

type Config struct {
	Lang string `json:"lang,omitempty"`
}

// Path returns the location of config.json.
func Path() (string, error) {
	dir, err := platform.DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// Load reads the settings. A missing or unreadable file yields defaults.
func Load() (Config, error) {
	p, err := Path()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return Config{}, nil
	}
	var c Config
	if json.Unmarshal(data, &c) != nil {
		return Config{}, nil
	}
	return c, nil
}

// Save writes the settings atomically.
func Save(c Config) error {
	p, err := Path()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(p), "config-*.tmp")
	if err != nil {
		return err
	}
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	if err := os.Rename(tmp.Name(), p); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return nil
}
