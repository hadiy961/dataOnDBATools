// File: internal/config/types.go
package config

import (
	"database/sql"
	"dbaTools/internal/utils"
)

// Config represents the main configuration structure
type Config struct {
	DB          *sql.DB
	AppSettings map[string]interface{}
	YAMLConfig  *utils.YAMLConfig
}

// Settings holds the loaded application settings
type Settings struct {
	AppName           string
	AppVersion        string
	AppEnv            string
	DebugMode         bool
	LogPath           string
	BackupPath        string
	BackupPrefix      string
	BackupEncryptKey  string
	BackupCompression bool
	BackupEncrypt     bool
	TempPath          string
}
