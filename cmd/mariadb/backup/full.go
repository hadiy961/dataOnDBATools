package MariaDBBackupCMD

import (
	"fmt"

	"github.com/spf13/cobra"
)

// cmd/config/zone/list.go
var MariaDBBackupFullCMD = &cobra.Command{
	Use:   "full",
	Short: "Full Backup MariaDB",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("FUll Backup mariaDB")
	},
}
