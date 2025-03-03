package main

import (
	"dbaTools/cmd"
	"dbaTools/internal/config"
	"dbaTools/internal/logger"
	"dbaTools/internal/utils"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	config, err := config.New()
	if err != nil {
		log.Fatal(err)
	}
	utils.ClearScreen()
	// Access settings
	settings := config.GetSettings()

	logger := logger.NewLogger(config.DB, settings.AppEnv, settings.AppName, settings.AppVersion, settings.DebugMode)
	logger.Info("Application " + settings.AppName + " v" + settings.AppVersion + " Started")
	// Store command name
	cmd.RootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		commandPath := cmd.CommandPath()
		flags := []string{}

		cmd.Flags().Visit(func(f *pflag.Flag) {
			flags = append(flags, "--"+f.Name+"="+f.Value.String())
		})

		if len(flags) > 0 {
			commandPath += " " + strings.Join(flags, " ")
		}

		logger.Info("Command [" + commandPath + "] executed")
	}

	// Execute root command
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
