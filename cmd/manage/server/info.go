package serverCMD

import (
	"dbaTools/internal/config"
	"dbaTools/internal/logger"
	"dbaTools/internal/package/manage/server"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var ListServerCMD = &cobra.Command{
	Use:   "info",
	Short: "Server information",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check if server ID flag was provided
		flagProvided := false
		cmd.Flags().Visit(func(f *pflag.Flag) {
			if f.Name == "id" {
				flagProvided = true
			}
		})

		if !flagProvided {
			return fmt.Errorf("required flag \"id\" not set")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.New()
		if err != nil {
			log.Fatal(err)
		}
		settings := config.GetSettings()

		logger := logger.NewLogger(config.DB, settings.AppEnv, settings.AppName, settings.AppVersion, settings.DebugMode)
		server.GetServerInfoByID(logger, config.DB, serverId)

	},
}

func init() {
	// You need to add "pflag" import if you don't already have it
	// import "github.com/spf13/pflag"

	ListServerCMD.Flags().IntVarP(&serverId, "id", "i", 0, "Server ID to check")
	ListServerCMD.MarkFlagRequired("id") // Mark the flag as required

}
