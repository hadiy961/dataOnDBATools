package lokal_rpm

import (
	"dbaTools/internal/logger"
	"dbaTools/internal/utils"
	"fmt"
)

// PackageDownloader interface mendefinisikan kontrak untuk download package
type PackageDownloader interface {
	PrepareDependencies() error
	ConfigureRepository() error
	DownloadPackages() error
}

// BaseDownloader berisi fungsi umum yang digunakan semua downloader
type BaseDownloader struct {
	packagesDir string
	osInfo      *utils.OSInfo
}

// NewBaseDownloader membuat instance baru BaseDownloader
func NewBaseDownloader(packagesDir string, osInfo *utils.OSInfo) *BaseDownloader {
	return &BaseDownloader{
		packagesDir: packagesDir,
		osInfo:      osInfo,
	}
}

// getDownloader returns the appropriate downloader based on OS
func getDownloader(packagesDir string, osInfo *utils.OSInfo, log *logger.Logger) (PackageDownloader, error) {
	baseDownloader := NewBaseDownloader(packagesDir, osInfo)

	switch osInfo.Distribution {
	case "ubuntu":
		return NewUbuntuDownloader(baseDownloader, log), nil
	case "centos", "red hat enterprise linux", "rocky linux":
		return NewCentOSDownloader(baseDownloader), nil
	default:
		return nil, fmt.Errorf("unsupported distribution: %s", osInfo.Distribution)
	}
}
