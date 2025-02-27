package SetupMariaDBCMD

import (
	"github.com/spf13/cobra"
)

var SetupOnlineCMD = &cobra.Command{
	Use:   "online",
	Short: "Setup MariaDB menggunakan internet",
	Run: func(cmd *cobra.Command, args []string) {

	},
}
