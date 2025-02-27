package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func NewMySQLDumper(config *MySQLDumpConfig) *MySQLDumper {
	return &MySQLDumper{
		config: config,
	}
}

type DumpProgress struct {
	StartTime    time.Time
	TotalBytes   int64
	CurrentDB    string
	CurrentTable string
}

func formatBytes(bytes int64) string {
	mb := float64(bytes) / 1024 / 1024
	if mb < 1024 {
		return fmt.Sprintf("%.0f MB", mb)
	}
	gb := mb / 1024
	return fmt.Sprintf("%.2f GB", gb)
}

func (d *MySQLDumper) buildBaseCommand() []string {
	cmd := []string{
		"mysqldump",
		"-h", d.config.Host,
		"-P", fmt.Sprintf("%d", d.config.Port),
		"-u", d.config.User,
		"-p" + d.config.Password,
		"--max-allowed-packet=1G",
		"--hex-blob",
		"--order-by-primary",
		"--single-transaction",
		"--routines=true",
		"--triggers=true",
		"--opt",
		"-CfQq",
		"--verbose",
		"--extended-insert",
		"--dump-date", // Added for better progress monitoring
	}

	if d.config.StructOnly {
		cmd = append(cmd, "--no-data")
	}

	return cmd
}

func (d *MySQLDumper) DumpFullDatabase() error {

	cmd := d.buildBaseCommand()
	cmd = append(cmd, "--all-databases")

	return d.executeDump(cmd, "full_backup")
}

func (d *MySQLDumper) DumpMultipleDatabases() error {
	if len(d.config.Databases) == 0 {
		return fmt.Errorf("no databases specified for multiple database dump")
	}

	cmd := d.buildBaseCommand()
	cmd = append(cmd, "--databases")
	cmd = append(cmd, d.config.Databases...)

	return d.executeDump(cmd, "multi_db_backup")
}

func (d *MySQLDumper) DumpSingleDatabase(dbName string) error {

	cmd := d.buildBaseCommand()

	if len(d.config.Tables) > 0 {
		cmd = append(cmd, dbName)
		cmd = append(cmd, d.config.Tables...)
	} else if len(d.config.ExcludeTabs) > 0 {
		cmd = append(cmd, dbName)
		for _, table := range d.config.ExcludeTabs {
			cmd = append(cmd, fmt.Sprintf("--ignore-table=%s.%s", dbName, table))
		}
	} else {
		cmd = append(cmd, dbName)
	}

	return d.executeDump(cmd, fmt.Sprintf("%s_backup", dbName))
}

func (d *MySQLDumper) DumpUsersAndGrants() error {

	cmd := []string{
		"mysqldump",
		"-h", d.config.Host,
		"-P", fmt.Sprintf("%d", d.config.Port),
		"-u", d.config.User,
		"-p" + d.config.Password,
		"--no-data",
		"mysql",
		"user",
		"db",
		"tables_priv",
		"columns_priv",
		"procs_priv",
	}

	return d.executeDump(cmd, "users_grants_backup")
}

func (d *MySQLDumper) executeDump(cmdArgs []string, filePrefix string) error {
	timestamp := time.Now().Format("20060102_150405")
	sqlFile := filepath.Join(d.config.OutputDir, fmt.Sprintf("%s_%s.sql", filePrefix, timestamp))

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)

	outFile, err := os.Create(sqlFile)
	if err != nil {
		return fmt.Errorf("failed to create SQL file: %w", err)
	}
	defer outFile.Close()

	cmd.Stdout = outFile
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start mysqldump: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("mysqldump failed: %w", err)
	}

	fmt.Printf("\rStarting compression...")
	if err := CompressFile(sqlFile); err != nil {
		return fmt.Errorf("compression failed: %w", err)
	}

	if err := os.Remove(sqlFile); err != nil {
		return fmt.Errorf("failed to remove original SQL file : %w", err)

	}

	fileInfo, err := os.Stat(sqlFile + ".gz")
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	fmt.Printf("\nBackup completed successfully\n")
	fmt.Printf("- File: %s.gz\n", sqlFile)
	fmt.Printf("- Size: %s\n", formatBytes(fileInfo.Size()))

	return nil
}

func (d *MySQLDumper) ValidateConfig() error {
	if d.config.Host == "" || d.config.Port == 0 || d.config.User == "" {
		return fmt.Errorf("invalid MySQL connection configuration")
	}

	if d.config.OutputDir == "" {
		return fmt.Errorf("output directory not specified")
	}

	// Create output directory if it doesn't exist
	dirConfig := NewDirectoryConfig(d.config.OutputDir)
	if err := CreateDirectory([]string{""}, dirConfig); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return nil
}
