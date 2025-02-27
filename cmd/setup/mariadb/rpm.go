package SetupMariaDBCMD

import (
	"dbaTools/internal/config"
	"dbaTools/internal/logger"
	lokalrpm "dbaTools/internal/package/setup/lokal_rpm"
	"log"

	"github.com/spf13/cobra"
)

var SetupRPMCmd = &cobra.Command{
	Use:   "rpm",
	Short: "Download MariaDB RPM untuk di transfer ke server tanpa internet",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.New()
		if err != nil {
			log.Fatal(err)
		}
		settings := config.GetSettings()

		logger := logger.NewLogger(config.DB, settings.AppEnv, settings.AppName, settings.AppVersion, settings.DebugMode)

		if err := lokalrpm.DownloadRPM(logger); err != nil {
			cmd.PrintErrf("Error: %v\n", err)
		} else {

			// Handle transfer prompts and process
			if err := lokalrpm.PromptForTransfer(logger); err != nil {
				cmd.PrintErrf("Error during transfer process: %v\n", err)
			}
		}

	},
}
