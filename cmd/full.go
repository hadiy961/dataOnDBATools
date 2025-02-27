package cmd

import (
	backup_full "dbaTools/internal/package/backup/full"
	"dbaTools/internal/utils"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	fullHost        string
	fullPort        int
	fullUsername    string
	fullPassword    string
	fullDestination string
)

var FullBackup = &cobra.Command{
	Use:   "full",
	Short: "Perform a complete backup of all databases",
	Long:  `Performs a full backup of all databases to the specified destination path.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create backup configuration
		backupConfig := &utils.BackupConfig{
			Host:        fullHost,
			Port:        fullPort,
			Username:    fullUsername,
			Password:    fullPassword,
			Destination: fullDestination,
		}

		// Get database information
		db, err := backup_full.ConnectDatabase(backupConfig)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
		defer db.Close()

		databases, err := backup_full.GetDatabaseInfo(db)
		if err != nil {
			return fmt.Errorf("failed to get database information: %w", err)
		}

		backup_full.ShowBackupSummary(backupConfig)

		// Display database information
		for dbName, size := range databases {
			fmt.Printf("Database: %s, Size: %s\n",
				dbName, utils.FormatBytes(uint64(size)))
		}

		// Confirm before proceeding
		if !utils.PromptYesNo("\nDo you want to proceed with the backup?", false) {
			fmt.Println("Backup cancelled")
			return nil
		}

		// Execute the backup
		if err := backup_full.ExecuteFullBackup(backupConfig); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	FullBackup.Flags().StringVarP(&fullHost, "host", "H", "", "MySQL server hostname")
	FullBackup.Flags().IntVarP(&fullPort, "port", "P", 0, "MySQL server port")
	FullBackup.Flags().StringVarP(&fullUsername, "username", "u", "", "MySQL username")
	FullBackup.Flags().StringVarP(&fullPassword, "password", "p", "", "MySQL password")
	FullBackup.Flags().StringVarP(&fullDestination, "destination", "o", "", "Backup destination path")

	FullBackup.MarkFlagRequired("host")
	FullBackup.MarkFlagRequired("port")
	FullBackup.MarkFlagRequired("username")
	FullBackup.MarkFlagRequired("password")
}
