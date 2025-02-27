package SetupMariaDBCMD

import (
	"dbaTools/internal/config"
	"dbaTools/internal/logger"
	"dbaTools/internal/package/setup/lokal"
	"log"

	"github.com/spf13/cobra"
)

var SetupLokalOfflineCMD = &cobra.Command{
	Use:   "lokal",
	Short: "Setup MariaDB menggunakan RPM (No Internet)",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.New()
		if err != nil {
			log.Fatal(err)
		}
		settings := config.GetSettings()

		logger := logger.NewLogger(config.DB, settings.AppEnv, settings.AppName, settings.AppVersion, settings.DebugMode)

		lokal.StartSetupLocal(logger)
	},
}
