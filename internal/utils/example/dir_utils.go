package example

import (
	"dbaTools/internal/utils"
	"fmt"
	"os"
	"syscall"
)

func ExampleUsage() {
	// Konfigurasi direktori
	config := utils.NewDirectoryConfig("/tmp/test")

	// 1. Membuat direktori
	err := utils.CreateDirectory([]string{"data", "logs"}, config)
	handleError(err, "Create directories")

	// 2. Buat nested direktori
	err = utils.CreateNestedDirectories("data/2024/01/logs", config)
	handleError(err, "Create nested directories")

	// 3. Cek dan tampilkan info direktori
	dirInfo, err := utils.CheckDirectory("/tmp/test/data")
	handleError(err, "Check directory")
	printDirectoryInfo(dirInfo)

	// 4. Copy direktori
	err = utils.CopyDirectory("/tmp/test/data", "/tmp/test/data_backup")
	handleError(err, "Copy directory")

	// 5. Cari file dalam direktori
	matches, err := utils.FindInDirectory("/tmp/test", "*.log")
	handleError(err, "Find in directory")
	fmt.Printf("Found files: %v\n", matches)

	// 6. Statistik direktori
	stats, err := utils.GetDirectoryStats("/tmp/test")
	handleError(err, "Get directory stats")
	printDirectoryStats(stats)

	// 7. Lock direktori
	lock, err := utils.LockDirectory("/tmp/test/data")
	handleError(err, "Lock directory")
	defer lock.Close()

	// 8. Cek disk usage
	usage, err := utils.GetDiskUsage("/tmp/test")
	handleError(err, "Get disk usage")
	printDiskUsage(usage)

	// 9. Dapatkan subdirektori
	subDirs, err := utils.GetSubDirectories("/tmp/test")
	handleError(err, "Get subdirectories")
	fmt.Printf("Subdirectories: %v\n", subDirs)

	// 10. Rename/pindah direktori
	err = utils.MoveDirectory("/tmp/test/data", "/tmp/test/data_new")
	handleError(err, "Move directory")

	// 11. Hapus direktori
	err = utils.RemoveDirectory("/tmp/test/data_new", true)
	handleError(err, "Remove directory")
}

func handleError(err error, operation string) {
	if err != nil {
		fmt.Printf("%s error: %v\n", operation, err)
		os.Exit(1)
	}
}

func printDirectoryInfo(info *utils.DirectoryInfo) {
	fmt.Printf("Directory Info:\n")
	fmt.Printf("  Path: %s\n", info.Path)
	fmt.Printf("  Exists: %v\n", info.Exists)
	fmt.Printf("  IsDir: %v\n", info.IsDir)
	fmt.Printf("  Mode: %v\n", info.Mode)
	fmt.Printf("  Size: %d bytes\n", info.Size)
	fmt.Printf("  LastMod: %v\n", info.LastMod)
	fmt.Printf("  IsEmpty: %v\n", info.IsEmpty)
	fmt.Printf("  Children: %v\n", info.Children)
}

func printDirectoryStats(stats *utils.DirectoryStats) {
	fmt.Printf("Directory Stats:\n")
	fmt.Printf("  Total Files: %d\n", stats.TotalFiles)
	fmt.Printf("  Total Dirs: %d\n", stats.TotalDirs)
	fmt.Printf("  Total Size: %d bytes\n", stats.TotalSize)
	fmt.Printf("  Last Modified: %v\n", stats.LastModified)
	fmt.Printf("  Max Depth: %d\n", stats.MaxDepth)
}

func printDiskUsage(usage *syscall.Statfs_t) {
	fmt.Printf("Disk Usage:\n")
	fmt.Printf("  Total: %d bytes\n", usage.Blocks*uint64(usage.Bsize))
	fmt.Printf("  Free: %d bytes\n", usage.Bfree*uint64(usage.Bsize))
	fmt.Printf("  Available: %d bytes\n", usage.Bavail*uint64(usage.Bsize))
}
