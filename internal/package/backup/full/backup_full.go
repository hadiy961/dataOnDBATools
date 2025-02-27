package backup_full

import (
	"dbaTools/internal/database"
	"dbaTools/internal/utils"
	"fmt"

	"github.com/dustin/go-humanize"
)

type BackupError struct {
	op  string
	err error
}

func (e *BackupError) Error() string {
	return fmt.Sprintf("%s: %v", e.op, e.err)
}

func ShowBackupSummary(cfg *utils.BackupConfig) {
	fmt.Println("======= Backup Summary =======")
	fmt.Printf("Host: %s\n", cfg.Host)
	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("Username: %s\n", cfg.Username)
	fmt.Printf("Destination: %s\n", cfg.Destination)
	fmt.Println("==============================")
}

func ValidateDestination(destination string) error {
	dirConfig := utils.NewDirectoryConfig(destination)
	if err := utils.CreateDirectory([]string{""}, dirConfig); err != nil {
		return &BackupError{"create_destination", err}
	}

	dirInfo, err := utils.CheckDirectory(destination)
	if err != nil {
		return &BackupError{"check_destination", err}
	}

	if !dirInfo.IsDir {
		return &BackupError{"invalid_destination", fmt.Errorf("destination path is not a directory")}
	}

	return nil
}

func ConnectDatabase(cfg *utils.BackupConfig) (*database.DBConnection, error) {
	dbConfig := database.DBConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.Username,
		Password: cfg.Password,
	}

	db := database.NewDBConnection(dbConfig)
	if err := db.Connect(); err != nil {
		return nil, &BackupError{"db_connect", err}
	}

	// Verify connection is working
	if err := db.CheckConnection(); err != nil {
		db.Close()
		return nil, &BackupError{"db_check", err}
	}

	return db, nil
}

func GetDatabaseInfo(db *database.DBConnection) (map[string]int64, error) {
	databases, err := db.GetDatabases()
	if err != nil {
		return nil, &BackupError{"get_databases", err}
	}

	if len(databases) == 0 {
		return nil, &BackupError{"empty_databases", fmt.Errorf("no databases found to backup")}
	}

	return databases, nil
}

func ExecuteFullBackup(cfg *utils.BackupConfig) error {
	// Validate destination directory
	if err := ValidateDestination(cfg.Destination); err != nil {
		return fmt.Errorf("destination validation failed: %w", err)
	}

	// Connect to database
	db, err := ConnectDatabase(cfg)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	defer db.Close()

	// Get database information
	databases, err := GetDatabaseInfo(db)
	if err != nil {
		return fmt.Errorf("failed to get database information: %w", err)
	}

	// Log database information
	for dbName, size := range databases {
		fmt.Printf("Database: %s, Size: %s\n", dbName, humanize.Bytes(uint64(size)))
	}

	if !utils.PromptYesNo("\nDo you want to proceed with the backup?", false) {
		fmt.Println("Backup cancelled")
		return nil
	}

	// Initialize and validate dumper
	dumpConfig := &utils.MySQLDumpConfig{
		Host:      cfg.Host,
		Port:      cfg.Port,
		User:      cfg.Username,
		Password:  cfg.Password,
		OutputDir: cfg.Destination,
	}

	dumper := utils.NewMySQLDumper(dumpConfig)
	if err := dumper.ValidateConfig(); err != nil {
		return fmt.Errorf("invalid backup configuration: %w", err)
	}

	// Execute backup
	fmt.Println("\nStarting full backup...")
	if err := dumper.DumpFullDatabase(); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	return nil
}
