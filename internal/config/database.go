// File: internal/config/database.go
package config

import (
	"database/sql"
	"dbaTools/internal/database"
	"dbaTools/internal/utils"
	"fmt"
)

func initializeDatabase(yamlConfig *utils.YAMLConfig) (*sql.DB, error) {
	dbConfig := database.DBConfig{
		Host:     yamlConfig.Database.Host,
		Port:     yamlConfig.Database.Port,
		User:     yamlConfig.Database.User,
		Password: yamlConfig.Database.Pass,
		Database: yamlConfig.Database.Name,
	}

	conn := database.NewDBConnection(dbConfig)
	if err := conn.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return conn.GetDB(), nil
}
