package lokal

import (
	"dbaTools/internal/logger"
	"fmt"
)

func handleRemoteSetup(log *logger.Logger) error {
	log.Info("Setup remote server belum tersedia")
	fmt.Println("\nMaaf, setup untuk remote server masih dalam pengembangan.")
	fmt.Println("Silakan gunakan opsi setup lokal untuk saat ini.")
	return nil
}
