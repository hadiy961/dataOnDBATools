package cmd

import (
	"dbaTools/cmd/manage"

	"github.com/spf13/cobra"
)

// cmd/config/config.go
var ManageRootCMD = &cobra.Command{
	Use:   "manage",
	Short: "Management",
}

func init() {
	ManageRootCMD.AddCommand(manage.ZoneCmd)
	ManageRootCMD.AddCommand(manage.StatusCMD)
	ManageRootCMD.AddCommand(manage.ServerCMD)
}
