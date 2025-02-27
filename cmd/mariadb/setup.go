package MariaDBCMD

import (
	MariaDBSetupCMD "dbaTools/cmd/mariadb/setup"

	"github.com/spf13/cobra"
)

// cmd/config/config.go
var MariaDBSetup = &cobra.Command{
	Use:   "setup",
	Short: "MariaDB Setup Command",
}

func init() {
	MariaDBSetup.AddCommand(MariaDBSetupCMD.MariaDBSetupHostingCMD)
	MariaDBSetup.AddCommand(MariaDBSetupCMD.MariaDBSetupOnPremiseCMD)
	MariaDBSetup.AddCommand(MariaDBSetupCMD.MariaDBSetupSaasCMD)
}
