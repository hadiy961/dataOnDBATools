package zoneCMD

import (
	"bufio"
	"dbaTools/internal/config"
	"dbaTools/internal/logger"
	"dbaTools/internal/package/manage/zone"
	"dbaTools/internal/utils"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// cmd/config/zone/add.go
var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new status",
	Run: func(cmd *cobra.Command, args []string) {

		config, err := config.New()
		if err != nil {
			log.Fatal(err)
		}
		settings := config.GetSettings()

		logger := logger.NewLogger(config.DB, settings.AppEnv, settings.AppName, settings.AppVersion, settings.DebugMode)

		errAdd := zone.AddZone(logger, config.DB)

		if errAdd == nil {
			// Tanyakan apakah ingin kembali ke daftar zona
			if utils.PromptYesNo("Return to zone list?", true) {
				zone.DisplayZoneList(logger, config.DB)
			}
		} else {
			// Tanyakan apakah ingin kembali ke daftar zona
			if utils.PromptYesNo("Return to zone list?", true) {
				zone.DisplayZoneList(logger, config.DB)
			} else {
				// Jika tidak, pastikan terminal dalam keadaan normal
				fmt.Println("Press enter to continue...")
				bufio.NewReader(os.Stdin).ReadBytes('\n')
			}
		}

	},
}
