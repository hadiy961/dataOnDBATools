package manage

import (
	statuscmd "dbaTools/cmd/manage/status"

	"github.com/spf13/cobra"
)

// cmd/config/Status/Status.go
var StatusCMD = &cobra.Command{
	Use:   "status",
	Short: "Status configuration commands",
}

func init() {
	// Add subcommands to Status
	StatusCMD.AddCommand(statuscmd.AddCmd, statuscmd.ListCmd, statuscmd.UpdateCmd, statuscmd.DeleteCMD)
}
