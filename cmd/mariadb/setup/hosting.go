package MariaDBSetupCMD

import (
	"github.com/spf13/cobra"
)

// cmd/config/zone/list.go
var MariaDBSetupHostingCMD = &cobra.Command{
	Use:   "hosting",
	Short: "Setup MariaDB untuk hosting",
	Run: func(cmd *cobra.Command, args []string) {

	},
}
