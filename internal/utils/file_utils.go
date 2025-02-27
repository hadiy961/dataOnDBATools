package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"os/user"
	"path"
	"regexp"
	"strings"
	"syscall"
	"time"
)

// CopyFile menyalin file dari sumber ke tujuan
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// WriteFile menulis data ke file
func WriteFile(filepath string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filepath, data, perm)
}

// RemoveFile menghapus file
func RemoveFile(filepath string) error {
	if !FileExists(filepath) {
		return fmt.Errorf("file tidak ditemukan: %s", filepath)
	}
	return os.Remove(filepath)
}

// RenameFileOS mengganti nama file menggunakan os.Rename
func RenameFileOS(oldpath, newpath string) error {
	if !FileExists(oldpath) {
		return fmt.Errorf("file sumber tidak ditemukan: %s", oldpath)
	}
	if FileExists(newpath) {
		return fmt.Errorf("file tujuan sudah ada: %s", newpath)
	}
	return os.Rename(oldpath, newpath)
}

// MoveFile memindahkan file dari sumber ke tujuan
func MoveFile(src, dst string) error {
	if err := CopyFile(src, dst); err != nil {
		return err
	}
	return os.Remove(src)
}

// RenameFile mengganti nama file
func RenameFile(old, new string) error {
	if !FileExists(old) {
		return fmt.Errorf("file sumber tidak ditemukan")
	}
	if FileExists(new) {
		return fmt.Errorf("file tujuan sudah ada")
	}
	return os.Rename(old, new)
}

// ValidateFileSize mengecek ukuran file
func ValidateFileSize(filepath string, minSize, maxSize int64) error {
	info, err := os.Stat(filepath)
	if err != nil {
		return err
	}

	size := info.Size()
	if size < minSize {
		return fmt.Errorf("ukuran file terlalu kecil (minimum %d bytes)", minSize)
	}
	if size > maxSize {
		return fmt.Errorf("ukuran file terlalu besar (maksimum %d bytes)", maxSize)
	}
	return nil
}

// ValidateFileType mengecek tipe file yang diizinkan
func ValidateFileType(filepath string, allowedTypes []string) error {
	mimeType, err := GetMimeType(filepath)
	if err != nil {
		return err
	}

	for _, allowed := range allowedTypes {
		if strings.HasPrefix(mimeType, allowed) {
			return nil
		}
	}
	return fmt.Errorf("tipe file tidak diizinkan: %s", mimeType)
}

// ValidateFilename mengecek nama file valid
func ValidateFilename(filename string) error {
	// Karakter yang dilarang dalam nama file
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*]`)
	if invalidChars.MatchString(filename) {
		return fmt.Errorf("nama file mengandung karakter tidak valid")
	}

	// Cek panjang nama file
	if len(filename) > 255 {
		return fmt.Errorf("nama file terlalu panjang")
	}

	return nil
}

// ReadFile membaca seluruh file
func ReadFile(filepath string) ([]byte, error) {
	if !FileExists(filepath) {
		return nil, fmt.Errorf("file tidak ditemukan: %s", filepath)
	}

	if !IsReadable(filepath) {
		return nil, fmt.Errorf("file tidak dapat dibaca: %s", filepath)
	}

	return os.ReadFile(filepath)
}

// EncryptFile mengenkripsi file dengan AES-256
func EncryptFile(src, dst string, key []byte) error {
	// Validasi panjang key
	if len(key) != 32 {
		return fmt.Errorf("panjang key harus 32 bytes")
	}

	plaintext, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// Buat IV random
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	// Enkripsi
	stream := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	stream.XORKeyStream(ciphertext, plaintext)

	// Simpan IV + ciphertext
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write(iv); err != nil {
		return err
	}
	_, err = f.Write(ciphertext)
	return err
}

// DecryptFile mendekripsi file yang dienkripsi dengan AES-256
func DecryptFile(src, dst string, key []byte) error {
	if len(key) != 32 {
		return fmt.Errorf("panjang key harus 32 bytes")
	}

	ciphertext, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	if len(ciphertext) < aes.BlockSize {
		return fmt.Errorf("file terenkripsi terlalu pendek")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return os.WriteFile(dst, plaintext, 0644)
}

// FileExists mengecek apakah file ada atau tidak
func FileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// GetFileExtension mendapatkan ekstensi file
func GetFileExtension(filePath string) string {
	return strings.ToLower(path.Ext(filePath))
}

// GetMimeType mendapatkan MIME type dari file
func GetMimeType(filePath string) (string, error) {
	if !FileExists(filePath) {
		return "", fmt.Errorf("file tidak ditemukan")
	}

	// Mendapatkan ekstensi file
	ext := GetFileExtension(filePath)

	// Coba deteksi MIME type berdasarkan ekstensi
	mimeType := mime.TypeByExtension(ext)
	if mimeType != "" {
		return mimeType, nil
	}

	// Jika tidak bisa mendeteksi dari ekstensi, baca beberapa byte pertama file
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Baca 512 byte pertama untuk deteksi
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Deteksi MIME type dari konten
	mimeType = http.DetectContentType(buffer[:n])
	if mimeType == "" {
		return "application/octet-stream", nil
	}

	return mimeType, nil
}

// CalculateMD5 menghitung MD5 hash dari file
func CalculateMD5(filepath string) (string, error) {
	if !FileExists(filepath) {
		return "", fmt.Errorf("file tidak ditemukan")
	}

	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// CalculateSHA256 menghitung SHA256 hash dari file
func CalculateSHA256(filepath string) (string, error) {
	if !FileExists(filepath) {
		return "", fmt.Errorf("file tidak ditemukan")
	}

	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// IsEmpty mengecek apakah file kosong
func IsEmpty(filepath string) (bool, error) {
	if !FileExists(filepath) {
		return false, fmt.Errorf("file tidak ditemukan")
	}

	info, err := os.Stat(filepath)
	if err != nil {
		return false, err
	}
	return info.Size() == 0, nil
}

// IsHidden mengecek apakah file tersembunyi (khusus Unix/Linux)
func IsHidden(filePath string) bool {
	filename := path.Base(filePath)
	return strings.HasPrefix(filename, ".")
}

// GetLastModified mendapatkan waktu modifikasi terakhir
func GetLastModified(filepath string) (time.Time, error) {
	if !FileExists(filepath) {
		return time.Time{}, fmt.Errorf("file tidak ditemukan")
	}

	info, err := os.Stat(filepath)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// IsOlderThan mengecek apakah file lebih lama dari durasi yang ditentukan
func IsOlderThan(filepath string, duration time.Duration) (bool, error) {
	if !FileExists(filepath) {
		return false, fmt.Errorf("file tidak ditemukan")
	}

	info, err := os.Stat(filepath)
	if err != nil {
		return false, err
	}

	return time.Since(info.ModTime()) > duration, nil
}

// CheckFile melakukan pengecekan detail terhadap file
func CheckFile(filepath string) (*FileInfo, error) {
	info := &FileInfo{
		Path: filepath,
	}

	if !FileExists(filepath) {
		info.Exists = false
		return info, nil
	}

	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return nil, fmt.Errorf("error checking file: %v", err)
	}

	info.Exists = true
	info.Size = fileInfo.Size()
	info.Permissions = fileInfo.Mode().String()
	info.IsDir = fileInfo.IsDir()
	info.ModTime = fileInfo.ModTime()
	info.Extension = GetFileExtension(filepath)

	// Get MIME type
	if mimeType, err := GetMimeType(filepath); err == nil {
		info.MimeType = mimeType
	}

	// Calculate hashes
	if md5Hash, err := CalculateMD5(filepath); err == nil {
		info.MD5Hash = md5Hash
	}
	if sha256Hash, err := CalculateSHA256(filepath); err == nil {
		info.SHA256Hash = sha256Hash
	}

	// Get owner and group info
	if stat, ok := fileInfo.Sys().(*syscall.Stat_t); ok {
		uid := stat.Uid
		gid := stat.Gid

		if user, err := user.LookupId(fmt.Sprintf("%d", uid)); err == nil {
			info.Owner = user.Username
		}
		if group, err := user.LookupGroupId(fmt.Sprintf("%d", gid)); err == nil {
			info.Group = group.Name
		}
	}

	return info, nil
}

// IsReadable mengecek apakah file bisa dibaca
func IsReadable(filepath string) bool {
	// Cek dulu apakah file ada
	if !FileExists(filepath) {
		return false
	}

	file, err := os.OpenFile(filepath, os.O_RDONLY, 0)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// IsWritable mengecek apakah file bisa ditulis
func IsWritable(filepath string) bool {
	// Cek dulu apakah file ada
	if !FileExists(filepath) {
		return false
	}

	file, err := os.OpenFile(filepath, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// IsExecutable mengecek apakah file bisa dieksekusi
func IsExecutable(filepath string) bool {
	// Cek dulu apakah file ada
	if !FileExists(filepath) {
		return false
	}

	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return false
	}
	return fileInfo.Mode()&0111 != 0
}

// PrintFileDetails mencetak semua informasi file
func PrintFileDetails(info *FileInfo) {
	fmt.Printf("File Path: %s\n", info.Path)
	fmt.Printf("Exists: %v\n", info.Exists)

	if !info.Exists {
		return
	}

	fmt.Printf("Size: %d bytes\n", info.Size)
	fmt.Printf("Permissions: %s\n", info.Permissions)
	fmt.Printf("Is Directory: %v\n", info.IsDir)
	fmt.Printf("Extension: %s\n", info.Extension)
	fmt.Printf("MIME Type: %s\n", info.MimeType)
	fmt.Printf("Last Modified: %v\n", info.ModTime)
	fmt.Printf("Owner: %s\n", info.Owner)
	fmt.Printf("Group: %s\n", info.Group)
	fmt.Printf("MD5: %s\n", info.MD5Hash)
	fmt.Printf("SHA256: %s\n", info.SHA256Hash)
	fmt.Printf("Is Hidden: %v\n", IsHidden(info.Path))

	if empty, err := IsEmpty(info.Path); err == nil {
		fmt.Printf("Is Empty: %v\n", empty)
	}

	fmt.Printf("Readable: %v\n", IsReadable(info.Path))
	fmt.Printf("Writable: %v\n", IsWritable(info.Path))
	fmt.Printf("Executable: %v\n", IsExecutable(info.Path))
}
