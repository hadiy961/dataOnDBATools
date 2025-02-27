package MariaDBBackupCMD

import (
	"fmt"

	"github.com/spf13/cobra"
)

// cmd/config/zone/list.go
var MariaDBBackupSingleCMD = &cobra.Command{
	Use:   "single",
	Short: "Single DB Backup MariaDB",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Single DB Backup mariaDB")
	},
}
