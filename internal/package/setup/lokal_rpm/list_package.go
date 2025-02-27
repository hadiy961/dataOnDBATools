package lokal_rpm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PackageInfo struct {
	Name string
	Size int64
}

// savePackageList saves the list of downloaded packages to a file
func savePackageList(packagesDir string) error {
	// Read directory contents
	entries, err := os.ReadDir(packagesDir)
	if err != nil {
		return fmt.Errorf("failed to read packages directory: %w", err)
	}

	// Prepare package list content
	var content strings.Builder
	content.WriteString("MariaDB Packages Download List\n")
	content.WriteString("============================\n")
	content.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Group packages by type
	var mariadbPkgs, libPkgs, otherPkgs []PackageInfo

	var totalSize int64
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("failed to get file info: %w", err)
		}

		pkg := PackageInfo{
			Name: entry.Name(),
			Size: info.Size(),
		}
		totalSize += info.Size()

		name := strings.ToLower(entry.Name())
		if strings.Contains(name, "mariadb") {
			mariadbPkgs = append(mariadbPkgs, pkg)
		} else if strings.HasPrefix(name, "lib") {
			libPkgs = append(libPkgs, pkg)
		} else {
			otherPkgs = append(otherPkgs, pkg)
		}
	}

	// Write MariaDB packages
	if len(mariadbPkgs) > 0 {
		content.WriteString("MariaDB Core Packages:\n")
		var categorySize int64
		for _, pkg := range mariadbPkgs {
			content.WriteString(fmt.Sprintf("- %-60s %s\n", pkg.Name, formatSize(pkg.Size)))
			categorySize += pkg.Size
		}
		content.WriteString(fmt.Sprintf("Total Core Package Size: %s\n", formatSize(categorySize)))
		content.WriteString("\n")
	}

	// Write library dependencies
	if len(libPkgs) > 0 {
		content.WriteString("Library Dependencies:\n")
		var categorySize int64
		for _, pkg := range libPkgs {
			content.WriteString(fmt.Sprintf("- %-60s %s\n", pkg.Name, formatSize(pkg.Size)))
			categorySize += pkg.Size
		}
		content.WriteString(fmt.Sprintf("Total Library Size: %s\n", formatSize(categorySize)))
		content.WriteString("\n")
	}

	// Write other packages
	if len(otherPkgs) > 0 {
		content.WriteString("Other Dependencies:\n")
		var categorySize int64
		for _, pkg := range otherPkgs {
			content.WriteString(fmt.Sprintf("- %-60s %s\n", pkg.Name, formatSize(pkg.Size)))
			categorySize += pkg.Size
		}
		content.WriteString(fmt.Sprintf("Total Other Package Size: %s\n", formatSize(categorySize)))
		content.WriteString("\n")
	}

	// Add summary
	totalCount := len(mariadbPkgs) + len(libPkgs) + len(otherPkgs)
	content.WriteString("Summary:\n")
	content.WriteString("========\n")
	content.WriteString(fmt.Sprintf("Total Packages: %d\n", totalCount))
	content.WriteString(fmt.Sprintf("Total MariaDB Packages: %d\n", len(mariadbPkgs)))
	content.WriteString(fmt.Sprintf("Total Library Packages: %d\n", len(libPkgs)))
	content.WriteString(fmt.Sprintf("Total Other Packages: %d\n", len(otherPkgs)))
	content.WriteString(fmt.Sprintf("Total Download Size: %s\n", formatSize(totalSize)))

	// Save to file
	listPath := filepath.Join(packagesDir, "mariadb_packages_rpm_list.txt")
	if err := os.WriteFile(listPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write package list: %w", err)
	}

	return nil
}

// formatSize formats a size in bytes to a human-readable string
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
