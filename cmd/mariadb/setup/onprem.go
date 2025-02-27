package MariaDBSetupCMD

import (
	"github.com/spf13/cobra"
)

// cmd/config/zone/list.go
var MariaDBSetupOnPremiseCMD = &cobra.Command{
	Use:   "onpremise",
	Short: "Setup MariaDB untuk OnPremise",
	Run: func(cmd *cobra.Command, args []string) {

	},
}
