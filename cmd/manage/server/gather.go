package serverCMD

import (
	"dbaTools/internal/config"
	"dbaTools/internal/logger"
	"dbaTools/internal/package/manage/server"
	"log"

	"github.com/spf13/cobra"
)

var GatherServerInfoCMD = &cobra.Command{
	Use:   "gather",
	Short: "Gather Server information",
	Run: func(cmd *cobra.Command, args []string) {

		config, err := config.New()
		if err != nil {
			log.Fatal(err)
		}
		settings := config.GetSettings()

		logger := logger.NewLogger(config.DB, settings.AppEnv, settings.AppName, settings.AppVersion, settings.DebugMode)

		server.GatherAllServer(logger, config.DB)

	},
}
