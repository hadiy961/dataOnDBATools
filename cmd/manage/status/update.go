package zoneCMD

import (
	"dbaTools/internal/config"
	"dbaTools/internal/logger"
	"dbaTools/internal/package/manage/zone"
	"log"

	"github.com/spf13/cobra"
)

// cmd/config/zone/add.go
var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update status",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.New()
		if err != nil {
			log.Fatal(err)
		}
		settings := config.GetSettings()

		logger := logger.NewLogger(config.DB, settings.AppEnv, settings.AppName, settings.AppVersion, settings.DebugMode)

		zone.UpdateZone(logger, config.DB, id)
	},
}

var id string

func init() {
	UpdateCmd.Flags().StringVarP(&id, "id", "i", "", "Zone ID to update (required)")
	UpdateCmd.MarkFlagRequired("id")
}
