// Package config provides functionality to manage the application's configuration.
// It includes methods to create, save, and parse configuration files.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	configName = "config.json"
)

// Config represents the structure of the configuration file.
type Config struct {
	PathToSecretKey string `json:"path_to_secret_key"`
	AddrToService   string `json:"address_to_service"`
	folderPath      string
}

// New creates a new Config instance. If the configuration file does not exist,
// it returns an empty Config instance without an error.
func New(folderPath string) (config *Config, err error) {
	config, err = parseConfig(folderPath)

	if errors.Is(err, os.ErrNotExist) {
		err = nil
	}

	if config == nil && err == nil {
		config = &Config{}
		config.folderPath = folderPath
	}

	return
}

// Save writes the Config instance to the configuration file in the directory.
func (c *Config) Save() (err error) {
	pathToConfig := filepath.Join(c.folderPath, configName)
	file, err := os.Create(pathToConfig)

	if err != nil {
		return fmt.Errorf("cannot create config file: %w", err)
	}

	defer func() {
		err = errors.Join(err, file.Close())
	}()

	je := json.NewEncoder(file)

	err = je.Encode(c)

	if err != nil {
		return fmt.Errorf("cannot encode config: %w", err)
	}

	return nil
}

func parseConfig(folderPath string) (config *Config, err error) {
	pathToConfig := filepath.Join(folderPath, configName)
	file, err := os.Open(pathToConfig)

	if err != nil {
		return nil, fmt.Errorf("cannot open config file: %w", err)
	}

	defer func() {
		err = errors.Join(err, file.Close())
	}()

	config = &Config{
		folderPath: folderPath,
	}

	jd := json.NewDecoder(file)

	err = jd.Decode(config)

	if err != nil {
		return nil, fmt.Errorf("cannot decode config: %w", err)
	}

	return config, nil
}
