package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "dbaTools",
	Short: "Database backup and management utility",
}

func init() {
	// Register commands
	RootCmd.AddCommand(MariaDBRootCMD)
	RootCmd.AddCommand(ManageRootCMD)
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
