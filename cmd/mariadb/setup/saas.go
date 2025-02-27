package MariaDBSetupCMD

import (
	"dbaTools/internal/config"
	"dbaTools/internal/logger"
	saas_lokal "dbaTools/internal/package/mariadb/setup/saas/lokal"
	saas_remote "dbaTools/internal/package/mariadb/setup/saas/remote"
	"log"

	"github.com/spf13/cobra"
)

var (
	isRemote bool
)

// cmd/config/zone/list.go
var MariaDBSetupSaasCMD = &cobra.Command{
	Use:   "saas",
	Short: "Setup MariaDB untuk OnPremise",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.New()
		if err != nil {
			log.Fatal(err)
		}
		settings := config.GetSettings()

		logger := logger.NewLogger(config.DB, settings.AppEnv, settings.AppName, settings.AppVersion, settings.DebugMode)

		if isRemote {
			saas_remote.Setup(logger)
		} else {
			saas_lokal.Setup(logger)
		}
	},
}

func init() {
	// Add --remote flag to the command
	MariaDBSetupSaasCMD.Flags().BoolVar(&isRemote, "remote", false, "Set up on a remote server instead of local machine")
}
