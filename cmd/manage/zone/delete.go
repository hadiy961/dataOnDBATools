package zoneCMD

import (
	"dbaTools/internal/config"
	"dbaTools/internal/logger"
	"dbaTools/internal/package/manage/zone"
	"log"

	"github.com/spf13/cobra"
)

var DeleteCMD = &cobra.Command{
	Use:   "delete",
	Short: "Delete Zone",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.New()
		if err != nil {
			log.Fatal(err)
		}
		settings := config.GetSettings()

		logger := logger.NewLogger(config.DB, settings.AppEnv, settings.AppName, settings.AppVersion, settings.DebugMode)

		zone.DeleteZone(logger, config.DB, idDel)
	},
}

var idDel string

func init() {
	DeleteCMD.Flags().StringVarP(&idDel, "id", "i", "", "Zone ID to update (required)")
	DeleteCMD.MarkFlagRequired("id")
}
