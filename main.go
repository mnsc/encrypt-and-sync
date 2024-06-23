package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type FileMetadata struct {
	OriginalPath string `json:"original_path"`
	Hash         string `json:"hash"`
	ModTime      int64  `json:"mod_time"`
}

func main() {
	// Define command line options
	syncFlag := flag.Bool("sync", false, "Sync files to OneDrive")
	restoreFlag := flag.Bool("restore", false, "Restore files from OneDrive")
	flag.Parse()

	if *syncFlag {
		syncFiles()
	} else if *restoreFlag {
		// Placeholder for restore logic
		fmt.Println("Restore functionality not implemented yet.")
	} else {
		printHelp()
	}
}

func syncFiles() {
	rootFolder := "C:\\Temp\\SourceTest"
	oneDriveFolder := "C:\\Temp\\OneDriveTest"
	metadataFile := filepath.Join(oneDriveFolder, "metadata.json")

	metadata := loadMetadata(metadataFile)
	metadataMap := createMetadataMap(metadata)

	newPhotosCount := 0
	skippedPhotosCount := 0
	copiedFilesCount := 0
	fileExtensionCount := make(map[string]int) // Map to track file extension counts
	updatedPhotos := make(map[string]string)   // Map to track updated photos and their new hashes

	err := filepath.Walk(rootFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// Read the file
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Create the destination path
			relativePath, err := filepath.Rel(rootFolder, path)
			if err != nil {
				return err
			}

			var destPath string
			if strings.ToLower(filepath.Ext(path)) == ".cr2" {
				// Compute the hash of the file contents
				hash := sha256.Sum256(data)
				hashString := hex.EncodeToString(hash[:])

				// Get the last modified time in epoch seconds
				modTime := info.ModTime().Unix()

				// Check if the file with the same path exists in metadata
				if metadataEntry, exists := metadataMap[relativePath]; exists {
					// If the file exists but has a new hash, add to updatedPhotos
					if metadataEntry.Hash != hashString {
						updatedPhotos[relativePath] = hashString
						// Handle the file encryption and writing
						handleFileEncryptionAndWriting(oneDriveFolder, relativePath, hashString, data, info)
						// Update metadata
						newMetadata := FileMetadata{
							OriginalPath: relativePath,
							Hash:         hashString,
							ModTime:      modTime,
						}
						metadata = append(metadata, newMetadata)
						metadataMap[relativePath] = newMetadata
						newPhotosCount++
					} else {
						skippedPhotosCount++
					}
				} else {
					// Handle the file encryption and writing
					handleFileEncryptionAndWriting(oneDriveFolder, relativePath, hashString, data, info)
					// Update metadata
					newMetadata := FileMetadata{
						OriginalPath: relativePath,
						Hash:         hashString,
						ModTime:      modTime,
					}
					metadata = append(metadata, newMetadata)
					metadataMap[relativePath] = newMetadata

					newPhotosCount++
				}
			} else {
				// Check if the archive bit is set
				if isArchiveBitSet(path) {
					// Just copy the file
					destPath = filepath.Join(oneDriveFolder, relativePath)

					// Ensure the destination directory exists
					destDir := filepath.Dir(destPath)
					if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
						return err
					}

					// Write the file to the destination
					if err := os.WriteFile(destPath, data, info.Mode()); err != nil {
						return err
					}

					// Reset the archive bit
					resetArchiveBit(path)

					copiedFilesCount++
					ext := strings.ToLower(filepath.Ext(path))
					fileExtensionCount[ext]++ // Update the count for the file extension
				}
			}

		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Save metadata at the end
	saveMetadata(metadataFile, metadata)

	// Create the summary
	summary := fmt.Sprintf("------------------------------------------\n")
	summary += fmt.Sprintf("New photos: %d\n", newPhotosCount)
	summary += fmt.Sprintf("Skipped photos: %d\n", skippedPhotosCount)
	summary += fmt.Sprintf("------------------------------------------\n")
	summary += fmt.Sprintf("Copied %d other files\n", copiedFilesCount)
	summary += "Types:\n"
	for ext, count := range fileExtensionCount {
		summary += fmt.Sprintf("  %s: %d\n", ext, count)
	}

	// Add updated photos to the summary
	if len(updatedPhotos) > 0 {
		summary += "------------------------------------------\n"
		summary += "Updated photos:\n"
		for photo, hash := range updatedPhotos {
			summary += fmt.Sprintf("  %s (new hash: %s)\n", photo, hash)
		}
	}

	// Print the summary to standard out
	fmt.Print(summary)

	// Write the summary to a timestamped file
	timestamp := time.Now().Format("20060102-150405")
	summaryFile := filepath.Join(oneDriveFolder, fmt.Sprintf("sync-summary-%s.txt", timestamp))
	if err := os.WriteFile(summaryFile, []byte(summary), 0644); err != nil {
		fmt.Printf("Error writing summary file: %v\n", err)
	}
}

func handleFileEncryptionAndWriting(oneDriveFolder, relativePath, hashString string, data []byte, info os.FileInfo) error {
	// Create the destination path with hash
	destPath := filepath.Join(oneDriveFolder, relativePath+"-"+hashString+".encr")

	// Encrypt the file (placeholder for actual encryption)
	encryptedData := encrypt(data)

	// Ensure the destination directory exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}

	// Write the encrypted file to the destination
	if err := os.WriteFile(destPath, encryptedData, info.Mode()); err != nil {
		return err
	}

	return nil
}

func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("  -sync    Sync files to OneDrive")
	fmt.Println("  -restore Restore files from OneDrive")
}

func encrypt(data []byte) []byte {
	// Placeholder for actual encryption logic
	// For now, just return the data as-is
	return data
}

func loadMetadata(filePath string) []FileMetadata {
	var metadata []FileMetadata
	data, err := os.ReadFile(filePath)
	if err == nil {
		json.Unmarshal(data, &metadata)
	}
	return metadata
}

func saveMetadata(filePath string, metadata []FileMetadata) {
	data, _ := json.Marshal(metadata)
	os.WriteFile(filePath, data, 0644)
}

func createMetadataMap(metadata []FileMetadata) map[string]FileMetadata {
	metadataMap := make(map[string]FileMetadata)
	for _, entry := range metadata {
		key := entry.OriginalPath
		metadataMap[key] = entry
	}
	return metadataMap
}

func isFileInMetadataMap(metadataMap map[string]FileMetadata, path, hash string) bool {
	key := path + "-" + hash
	_, exists := metadataMap[key]
	return exists
}

func isArchiveBitSet(path string) bool {
	// Get file attributes
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	attrs, err := syscall.GetFileAttributes(pathPtr)
	if err != nil {
		return false
	}
	// Check if the archive bit is set
	return attrs&syscall.FILE_ATTRIBUTE_ARCHIVE != 0
}

func resetArchiveBit(path string) error {
	// Get current file attributes
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	attrs, err := syscall.GetFileAttributes(pathPtr)
	if err != nil {
		return err
	}
	// Remove the archive bit
	newAttrs := attrs &^ syscall.FILE_ATTRIBUTE_ARCHIVE
	return syscall.SetFileAttributes(pathPtr, newAttrs)
}
