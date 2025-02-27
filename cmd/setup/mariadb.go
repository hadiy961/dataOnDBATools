package setupCMD

import (
	SetupMariaDBCMD "dbaTools/cmd/setup/mariadb"

	"github.com/spf13/cobra"
)

// cmd/config/config.go
var MariaDBCMD = &cobra.Command{
	Use:   "setup",
	Short: "Setup MariaDB",
}

func init() {
	MariaDBCMD.AddCommand(SetupMariaDBCMD.SetupLokalCMD)
	MariaDBCMD.AddCommand(SetupMariaDBCMD.SetupLokalOfflineCMD)
	MariaDBCMD.AddCommand(SetupMariaDBCMD.SetupOnlineCMD)
	MariaDBCMD.AddCommand(SetupMariaDBCMD.SetupRPMCmd)
}
