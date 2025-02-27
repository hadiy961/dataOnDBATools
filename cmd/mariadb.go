package cmd

import (
	MariaDBCMD "dbaTools/cmd/mariadb"

	"github.com/spf13/cobra"
)

// cmd/config/config.go
var MariaDBRootCMD = &cobra.Command{
	Use:   "mariadb",
	Short: "MariaDB Command Tools",
}

func init() {
	MariaDBRootCMD.AddCommand(MariaDBCMD.MariaDBBackup)
	MariaDBRootCMD.AddCommand(MariaDBCMD.MariaDBSetup)
	MariaDBRootCMD.AddCommand(MariaDBCMD.MariaDBInstallCmd)
}
