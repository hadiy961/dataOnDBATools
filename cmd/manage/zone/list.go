package zoneCMD

import (
	"dbaTools/internal/config"
	"dbaTools/internal/logger"
	"dbaTools/internal/package/manage/zone"
	"log"

	"github.com/spf13/cobra"
)

// cmd/config/zone/list.go
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all zones",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.New()
		if err != nil {
			log.Fatal(err)
		}
		settings := config.GetSettings()

		logger := logger.NewLogger(config.DB, settings.AppEnv, settings.AppName, settings.AppVersion, settings.DebugMode)

		zone.DisplayZoneList(logger, config.DB)
	},
}
