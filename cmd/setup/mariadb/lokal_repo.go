package SetupMariaDBCMD

import (
	"github.com/spf13/cobra"
)

// cmd/config/zone/list.go
var SetupLokalCMD = &cobra.Command{
	Use:   "lokal_repo",
	Short: "Setup MariaDB menggunakan local repo (No Internet)",
	Run: func(cmd *cobra.Command, args []string) {

	},
}
