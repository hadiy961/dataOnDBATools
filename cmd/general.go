package cmd

import (
	"github.com/spf13/cobra"
)

var GeneralCMD = &cobra.Command{
	Use:   "general",
	Short: "Database backup and management utility",
}
