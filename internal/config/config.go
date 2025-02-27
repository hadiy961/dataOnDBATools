// File: internal/config/config.go
package config

import (
	"dbaTools/internal/utils"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// New initializes the configuration
func New() (*Config, error) {
	// Check and create config directory if needed
	configPath := filepath.Join(configDir, configFile)
	if err := ensureConfigFile(configPath); err != nil {
		return nil, err
	}

	// Read and parse YAML config
	yamlConfig, err := readYAMLConfig(configPath)
	if err != nil {
		return nil, err
	}

	// Validate YAML configuration
	if err := validateYAMLConfig(yamlConfig); err != nil {
		return nil, err
	}

	// Initialize database connection
	db, err := initializeDatabase(yamlConfig)
	if err != nil {
		return nil, err
	}

	// Load settings from database
	settings, err := loadSettings(db)
	if err != nil {
		return nil, err
	}

	config := &Config{
		DB:          db,
		AppSettings: settings,
		YAMLConfig:  yamlConfig,
	}

	// Validate and create required paths
	if err := validatePaths(config.GetSettings()); err != nil {
		return nil, err
	}

	return config, nil
}

func ensureConfigFile(configPath string) error {
	// Create config directory if it doesn't exist
	dirConfig := utils.NewDirectoryConfig(filepath.Dir(configPath))
	if err := utils.CreateDirectory([]string{""}, dirConfig); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if config file exists
	if !utils.FileExists(configPath) {
		// Create default config
		defaultConfig := utils.YAMLConfig{
			Database: struct {
				Host string `yaml:"db_host"`
				User string `yaml:"db_user"`
				Pass string `yaml:"db_pass"`
				Port int    `yaml:"db_port"`
				Name string `yaml:"db_name"`
			}{
				Host: "192.168.100.80",
				Port: 3306,
				User: "sst_user",
				Pass: "demo",
				Name: "dbsf_dba_tools",
			},
		}

		// Marshal config to YAML
		yamlData, err := yaml.Marshal(&defaultConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal default config: %w", err)
		}

		// Write config file
		if err := utils.WriteFile(configPath, yamlData, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}
	}

	return nil
}

func readYAMLConfig(configPath string) (*utils.YAMLConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config utils.YAMLConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GetSettings returns the application settings
func (c *Config) GetSettings() *Settings {
	return &Settings{
		AppName:           c.AppSettings["APP_NAME"].(string),
		AppVersion:        c.AppSettings["APP_VERSION"].(string),
		AppEnv:            c.AppSettings["APP_ENV"].(string),
		DebugMode:         c.AppSettings["DEBUG_MODE"].(bool),
		LogPath:           c.AppSettings["LOG_PATH"].(string),
		BackupPath:        c.AppSettings["BACKUP_PATH"].(string),
		BackupPrefix:      c.AppSettings["BACKUP_PREFIX"].(string),
		BackupEncryptKey:  c.AppSettings["BACKUP_ENCRYPT_KEY"].(string),
		BackupCompression: c.AppSettings["BACKUP_COMPRESSION"].(bool),
		BackupEncrypt:     c.AppSettings["BACKUP_ENCRYPT"].(bool),
		TempPath:          c.AppSettings["TEMP_PATH"].(string),
	}
}
