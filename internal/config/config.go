package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DBUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configPath := filepath.Join(homeDir, configFileName)
	return configPath, nil
}

func write(cfg Config) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, jsonData, 0644)
}

func (c *Config) SetUser(userName string) error {
	c.CurrentUserName = userName

	return write(*c)
}

func Read() (*Config, error) {
	configPath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
