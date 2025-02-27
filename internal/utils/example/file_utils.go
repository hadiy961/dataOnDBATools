package example

import (
	"crypto/rand"
	"dbaTools/internal/utils"
	"fmt"
	"os"
	"time"
)

// ExamplesAll menampilkan contoh penggunaan semua fungsi file utils
func ExamplesAll() {
	// Buat file contoh untuk testing
	testFile := "example.txt"
	content := []byte("Hello, this is a test file!")
	err := utils.WriteFile(testFile, content, 0644)
	if err != nil {
		fmt.Printf("Error creating test file: %v\n", err)
		return
	}
	defer utils.RemoveFile(testFile)

	// Buat file kedua untuk testing
	testFile2 := "example2.txt"
	err = utils.WriteFile(testFile2, content, 0644)
	if err != nil {
		fmt.Printf("Error creating second test file: %v\n", err)
		return
	}
	defer utils.RemoveFile(testFile2)

	fmt.Println("\n=== File Management Examples ===")

	// Copy file
	fmt.Println("\n1. CopyFile")
	err = utils.CopyFile(testFile, "copy.txt")
	if err != nil {
		fmt.Printf("Error copying file: %v\n", err)
	} else {
		fmt.Println("File copied successfully")
		os.Remove("copy.txt")
	}

	// Move file
	fmt.Println("\n2. MoveFile")
	err = utils.MoveFile(testFile2, "moved.txt")
	if err != nil {
		fmt.Printf("Error moving file: %v\n", err)
	} else {
		fmt.Println("File moved successfully")
		os.Remove("moved.txt")
	}

	// Rename file
	fmt.Println("\n3. RenameFile")
	err = utils.RenameFile(testFile, "renamed.txt")
	if err != nil {
		fmt.Printf("Error renaming file: %v\n", err)
	} else {
		fmt.Println("File renamed successfully")
		os.Rename("renamed.txt", testFile)
	}

	fmt.Println("\n=== File Validation Examples ===")

	// Validate file size
	fmt.Println("\n4. ValidateFileSize")
	err = utils.ValidateFileSize(testFile, 0, 1000)
	if err != nil {
		fmt.Printf("File size validation error: %v\n", err)
	} else {
		fmt.Println("File size validation passed")
	}

	// Validate file type
	fmt.Println("\n5. ValidateFileType")
	allowedTypes := []string{"text/", "application/"}
	err = utils.ValidateFileType(testFile, allowedTypes)
	if err != nil {
		fmt.Printf("File type validation error: %v\n", err)
	} else {
		fmt.Println("File type validation passed")
	}

	// Validate filename
	fmt.Println("\n6. ValidateFilename")
	err = utils.ValidateFilename("valid-file.txt")
	if err != nil {
		fmt.Printf("Filename validation error: %v\n", err)
	} else {
		fmt.Println("Filename validation passed")
	}

	fmt.Println("\n=== Encryption Examples ===")

	// Generate random key for encryption
	key := make([]byte, 32)
	rand.Read(key)

	// Encrypt file
	fmt.Println("\n7. EncryptFile")
	err = utils.EncryptFile(testFile, "encrypted.bin", key)
	if err != nil {
		fmt.Printf("Error encrypting file: %v\n", err)
	} else {
		fmt.Println("File encrypted successfully")
	}

	// Decrypt file
	fmt.Println("\n8. DecryptFile")
	err = utils.DecryptFile("encrypted.bin", "decrypted.txt", key)
	if err != nil {
		fmt.Printf("Error decrypting file: %v\n", err)
	} else {
		fmt.Println("File decrypted successfully")
		os.Remove("encrypted.bin")
		os.Remove("decrypted.txt")
	}

	fmt.Println("\n=== Compression Examples ===")

	// Compress file
	fmt.Println("\n9. CompressFile")
	err = utils.CompressFile("compressed.sql")
	if err != nil {
		fmt.Printf("Error compressing file: %v\n", err)
	} else {
		fmt.Println("File compressed successfully")
	}

	// Decompress file
	fmt.Println("\n10. DecompressFile")
	err = utils.DecompressFile("compressed.sql")
	if err != nil {
		fmt.Printf("Error decompressing file: %v\n", err)
	} else {
		fmt.Println("File decompressed successfully")
		os.Remove("compressed.gz")
		os.Remove("decompressed.txt")
	}

	fmt.Println("\n=== File Information Examples ===")

	// Get file info
	fmt.Println("\n11. GetFileInfo")
	if fileInfo, err := utils.CheckFile(testFile); err == nil {
		utils.PrintFileDetails(fileInfo)
	}

	// Check file existence
	fmt.Println("\n12. FileExists")
	fmt.Printf("File exists: %v\n", utils.FileExists(testFile))
	fmt.Printf("Non-existent file exists: %v\n", utils.FileExists("nonexistent.txt"))

	// Get file extension
	fmt.Println("\n13. GetFileExtension")
	fmt.Printf("Extension: %s\n", utils.GetFileExtension(testFile))

	// Get MIME type
	fmt.Println("\n14. GetMimeType")
	if mimeType, err := utils.GetMimeType(testFile); err == nil {
		fmt.Printf("MIME Type: %s\n", mimeType)
	}

	// Calculate hashes
	fmt.Println("\n15. Calculate Hashes")
	if md5Hash, err := utils.CalculateMD5(testFile); err == nil {
		fmt.Printf("MD5: %s\n", md5Hash)
	}
	if sha256Hash, err := utils.CalculateSHA256(testFile); err == nil {
		fmt.Printf("SHA256: %s\n", sha256Hash)
	}

	fmt.Println("\n=== File Status Examples ===")

	// Check if empty
	fmt.Println("\n16. IsEmpty")
	if isEmpty, err := utils.IsEmpty(testFile); err == nil {
		fmt.Printf("Is empty: %v\n", isEmpty)
	}

	// Check if hidden
	fmt.Println("\n17. IsHidden")
	fmt.Printf("Is hidden: %v\n", utils.IsHidden(testFile))

	// Get last modified time
	fmt.Println("\n18. GetLastModified")
	if modTime, err := utils.GetLastModified(testFile); err == nil {
		fmt.Printf("Last modified: %v\n", modTime)
	}

	// Check if older than
	fmt.Println("\n19. IsOlderThan")
	if isOld, err := utils.IsOlderThan(testFile, 1*time.Hour); err == nil {
		fmt.Printf("Is older than 1 hour: %v\n", isOld)
	}

	fmt.Println("\n=== File Permissions Examples ===")

	// Check permissions
	fmt.Println("\n20. Check Permissions")
	fmt.Printf("Is readable: %v\n", utils.IsReadable(testFile))
	fmt.Printf("Is writable: %v\n", utils.IsWritable(testFile))
	fmt.Printf("Is executable: %v\n", utils.IsExecutable(testFile))
}
