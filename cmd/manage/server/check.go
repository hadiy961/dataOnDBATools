package serverCMD

import (
	"dbaTools/internal/config"
	"dbaTools/internal/logger"
	"dbaTools/internal/package/manage/server"
	"log"

	"github.com/spf13/cobra"
)

var serverId int

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

		if serverId == 0 {
			server.GatherAllServer(logger, config.DB)
		} else {
			server.GatherServerByID(logger, config.DB, serverId)

		}
	},
}

func init() {
	GatherServerInfoCMD.Flags().IntVarP(&serverId, "id", "i", 0, "Server ID to check (leave empty for all servers)")
}
