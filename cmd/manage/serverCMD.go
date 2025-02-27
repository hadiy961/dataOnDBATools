package manage

import (
	serverCMD "dbaTools/cmd/manage/server"

	"github.com/spf13/cobra"
)

// cmd/config/zone/zone.go
var ServerCMD = &cobra.Command{
	Use:   "server",
	Short: "server manage commands",
}

func init() {
	// Add subcommands to zone
	ServerCMD.AddCommand(serverCMD.GatherServerInfoCMD, serverCMD.ListServerCMD)
}
