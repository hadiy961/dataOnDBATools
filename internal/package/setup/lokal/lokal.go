package lokal

import (
    "dbaTools/internal/logger"
    "dbaTools/internal/utils"
    "fmt"
)

func handleLocalSetup(log *logger.Logger) error {
    log.Info("Memulai setup server lokal")

    // Check system requirements
    if err := utils.CheckSystemRequirements(); err != nil {
        log.Error("Gagal memenuhi persyaratan sistem", err)
        return fmt.Errorf("system requirements check failed: %w", err)
    }

    // Get system info for logging
    sysInfo, err := utils.GetSystemInfo()
    if err != nil {
        log.Error("Gagal mendapatkan informasi sistem", err)
        return fmt.Errorf("failed to get system info: %w", err)
    }

    log.Info(fmt.Sprintf("Sistem terdeteksi: %s %s", sysInfo.Distribution, sysInfo.Version))

    // Confirm proceeding with installation
    if !utils.PromptYesNo("Lanjutkan instalasi?", true) {
        log.Info("Instalasi dibatalkan oleh pengguna")
        return nil
    }

    // Prompt for package location
    pkgPath, err := PromptPackageLocation("mariadb_packages")
    if err != nil {
        log.Error("Gagal mendapatkan lokasi package", err)
        return fmt.Errorf("failed to get package location: %w", err)
    }
    
    log.Info(fmt.Sprintf("Menggunakan package dari: %s", pkgPath))

    // Install packages
    if err := InstallPackages(pkgPath, log); err != nil {
        log.Error("Instalasi package gagal", err)
        return fmt.Errorf("package installation failed: %w", err)
    }

    return nil
}