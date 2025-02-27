package manage

import (
	zoneCMD "dbaTools/cmd/manage/zone"

	"github.com/spf13/cobra"
)

// cmd/config/zone/zone.go
var ZoneCmd = &cobra.Command{
	Use:   "zone",
	Short: "Zone configuration commands",
}

func init() {
	// Add subcommands to zone
	ZoneCmd.AddCommand(zoneCMD.AddCmd, zoneCMD.ListCmd, zoneCMD.UpdateCmd, zoneCMD.DeleteCMD)
}
