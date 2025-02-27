package lokal

import (
	"dbaTools/internal/logger"
	"dbaTools/internal/utils"
	"fmt"
)

// StartSetupLocal handles the local server setup process
func StartSetupLocal(log *logger.Logger) error {
	// Display welcome message
	fmt.Println("\nSelamat datang di Setup Wizard MariaDB")
	fmt.Println("=====================================")

	// Prompt for setup type
	setupType, err := promptSetupType()
	if err != nil {
		return fmt.Errorf("error during setup type selection: %w", err)
	}

	// Process based on selection
	switch setupType {
	case LocalSetup:
		return handleLocalSetup(log)
	case RemoteSetup:
		return handleRemoteSetup(log)
	default:
		return fmt.Errorf("pilihan tidak valid")
	}
}

// promptSetupType prompts the user to select the setup type
func promptSetupType() (SetupType, error) {
	fmt.Println("Server mana yang akan anda setup?")
	fmt.Println("1. Lokal (Server ini)")
	fmt.Println("2. Remote (Server lain)")

	choice, err := utils.PromptInt("Tentukan pilihan anda", 1, 1, 2)
	if err != nil {
		return 0, fmt.Errorf("error reading choice: %w", err)
	}

	return SetupType(choice), nil
}
