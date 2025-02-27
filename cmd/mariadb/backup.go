package MariaDBCMD

import (
	MariaDBBackupCMD "dbaTools/cmd/mariadb/backup"

	"github.com/spf13/cobra"
)

// cmd/config/config.go
var MariaDBBackup = &cobra.Command{
	Use:   "backup",
	Short: "MariaDB backup Command",
}

func init() {
	MariaDBBackup.AddCommand(MariaDBBackupCMD.MariaDBBackupFullCMD)
	MariaDBBackup.AddCommand(MariaDBBackupCMD.MariaDBBackupSingleCMD)
}
