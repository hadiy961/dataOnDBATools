package lokal_rpm

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// CentOSDownloader implementasi untuk CentOS
type CentOSDownloader struct {
	*BaseDownloader
}

// NewCentOSDownloader membuat instance baru CentOSDownloader
func NewCentOSDownloader(baseDownloader *BaseDownloader) *CentOSDownloader {
	return &CentOSDownloader{
		BaseDownloader: baseDownloader,
	}
}

func (c *CentOSDownloader) PrepareDependencies() error {
	// First ensure the directory exists
	if err := os.MkdirAll(c.packagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Get absolute path
	absPath, err := filepath.Abs(c.packagesDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}
	c.packagesDir = absPath

	// Check if dnf-utils is installed
	checkCmd := exec.Command("rpm", "-q", "dnf-utils")
	if err := checkCmd.Run(); err != nil {
		fmt.Println("Installing dnf-utils...")
		installCmd := exec.Command("dnf", "install", "-y", "dnf-utils")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to install dnf-utils: %w", err)
		}
	}
	return nil
}

func (c *CentOSDownloader) ConfigureRepository() error {
	repoCmd := exec.Command("bash", "-c", `cat > /etc/yum.repos.d/mariadb.repo << EOF
[mariadb]
name = MariaDB
baseurl = https://mirrors.aliyun.com/mariadb/yum/10.6/centos9-amd64
gpgkey = https://mirrors.aliyun.com/mariadb/yum/RPM-GPG-KEY-MariaDB
gpgcheck = 1
EOF`)

	repoCmd.Stdout = os.Stdout
	repoCmd.Stderr = os.Stderr
	if err := repoCmd.Run(); err != nil {
		return fmt.Errorf("failed to create MariaDB repository file: %w", err)
	}
	return nil
}

func (c *CentOSDownloader) DownloadPackages() error {
	packages := []string{
		"MariaDB-server",
		"MariaDB-client",
		"MariaDB-common",
		"MariaDB-backup",
		"MariaDB-shared",
	}

	fmt.Println("\nDownloading MariaDB packages...")
	for _, pkg := range packages {
		fmt.Printf("Downloading %s...\n", pkg)
		cmd := exec.Command("dnf", "download", "--resolve", "--alldeps", pkg)
		cmd.Dir = c.packagesDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: Failed to download %s: %v\n", pkg, err)
			continue
		}
	}

	fmt.Println("\nDownloading additional dependencies...")
	depCmd := exec.Command("dnf", "download", "--resolve", "--alldeps",
		"libaio",
		"ncurses-libs",
		"openssl-libs",
		"zlib")
	depCmd.Dir = c.packagesDir
	depCmd.Stdout = os.Stdout
	depCmd.Stderr = os.Stderr

	if err := depCmd.Run(); err != nil {
		fmt.Printf("Warning: Some dependencies may have failed to download: %v\n", err)
	}

	return nil
}
