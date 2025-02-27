package MariaDBCMD

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Available destination options
const (
	DestinationLocal  = "local"
	DestinationRemote = "remote"
)

var (
	// Destination flag
	destination string
)

// MariaDBInstallCmd represents the install command for MariaDB
var MariaDBInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install MariaDB database server",
	Long:  `This command installs MariaDB database server either locally or on a remote machine.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validate destination parameter
		if destination != DestinationLocal && destination != DestinationRemote {
			return fmt.Errorf("invalid destination: %s. Must be either 'local' or 'remote'", destination)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Starting MariaDB installation for destination: %s\n", destination)

		// Display installation options menu
		if destination == DestinationLocal {
			fmt.Println("Local installation selected")
		} else {
			fmt.Println("Remote installation selected")
		}
	},
}

func init() {
	// Add required destination flag
	MariaDBInstallCmd.Flags().StringVarP(&destination, "destination", "d", "",
		"Installation destination (required, must be either 'local' or 'remote')")
	MariaDBInstallCmd.MarkFlagRequired("destination")

}
