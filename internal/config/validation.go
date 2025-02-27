
// File: internal/config/validation.go
package config

import (
	"fmt"
	"dbaTools/internal/utils"
	"path/filepath"
)

const (
	configDir  = "configs"
	configFile = "dbaTools.yaml"
)

func validatePaths(settings *Settings) error {
	// Create required directories using existing utility functions
	paths := []string{settings.LogPath, settings.BackupPath, settings.TempPath}
	
	for _, path := range paths {
		if path == "" {
			return fmt.Errorf("path cannot be empty")
		}

		// Clean and split the path into segments
		cleanPath := filepath.Clean(path)
		if err := utils.CreateNestedDirectories(cleanPath, &utils.DirectoryConfig{
			BasePath:      "",
			Mode:         0755,
			CreateParents: true,
			SkipIfExists:  true,
		}); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
	}
	
	return nil
}

func validateYAMLConfig(cfg *utils.YAMLConfig) error {
	if cfg.Database.Host == "" || cfg.Database.Port == 0 || 
	   cfg.Database.User == "" || cfg.Database.Name == "" {
		return fmt.Errorf("invalid database configuration")
	}
	return nil
}
