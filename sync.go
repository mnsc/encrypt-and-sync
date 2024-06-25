package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

func syncFiles(sourceFolder, oneDriveFolder string, encrypt bool, key []byte, pathRegexp string) {
	startTime := time.Now() // Start timing the function

	metadataFile := filepath.Join(oneDriveFolder, "metadata.json")

	metadata := loadMetadata(metadataFile)
	metadataMap := createMetadataMap(metadata)

	newMediaFilesCount := 0
	skippedMediaFilesCount := 0
	copiedFilesCount := 0
	fileExtensionCount := make(map[string]int)   // Map to track file extension counts
	updatedMediaFiles := make(map[string]string) // Map to track updated media files and their new hashes

	mediaFileProcessingTime := time.Duration(0) // To track total time for new or updated media files

	// Compile the regular expression
	re, err := regexp.Compile(pathRegexp)
	if err != nil {
		fmt.Printf("Invalid regular expression: %v\n", err)
		return
	}

	err = filepath.Walk(sourceFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// Create the relative path
			relativePath, err := filepath.Rel(sourceFolder, path)
			if err != nil {
				return err
			}

			// Check if the relative path matches the regular expression
			if !re.MatchString(relativePath) {
				return nil
			}

			// Read the file
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var destPath string
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".cr2" || ext == ".jpg" || ext == ".mov" || ext == ".avi" {
				mediaFileStartTime := time.Now() // Start timing for this media file

				// Compute the hash of the file contents
				hash := sha256.Sum256(data)
				hashString := hex.EncodeToString(hash[:])
				fmt.Printf("#")

				// Get the last modified time in epoch seconds
				modTime := info.ModTime().Unix()

				// Check if the file with the same path exists in metadata
				if metadataEntry, exists := metadataMap[relativePath]; exists {
					// If the file exists but has a new hash, add to updatedMediaFiles
					if metadataEntry.Hash != hashString {
						updatedMediaFiles[relativePath] = hashString
						// Handle the file encryption and writing
						handleFileEncryptionAndWriting(oneDriveFolder, relativePath, hashString, data, key, info, encrypt)
						// Update metadata
						newMetadata := FileMetadata{
							OriginalPath: relativePath,
							Hash:         hashString,
							ModTime:      modTime,
						}
						metadata = append(metadata, newMetadata)
						metadataMap[relativePath] = newMetadata
						fmt.Printf("+")
						newMediaFilesCount++
						mediaFileProcessingTime += time.Since(mediaFileStartTime) // Add time taken for this media file
					} else {
						fmt.Printf(".")
						skippedMediaFilesCount++
					}
				} else {
					// Handle the file encryption and writing
					handleFileEncryptionAndWriting(oneDriveFolder, relativePath, hashString, data, key, info, encrypt)
					// Update metadata
					newMetadata := FileMetadata{
						OriginalPath: relativePath,
						Hash:         hashString,
						ModTime:      modTime,
					}
					metadata = append(metadata, newMetadata)
					metadataMap[relativePath] = newMetadata

					fmt.Printf("+")
					newMediaFilesCount++
					mediaFileProcessingTime += time.Since(mediaFileStartTime) // Add time taken for this media file
				}
			} else {
				// Create the destination path
				destPath = filepath.Join(oneDriveFolder, relativePath)

				// Check if the destination file exists
				if _, err := os.Stat(destPath); err == nil {
					// File exists, check if the archive bit is set
					if !isArchiveBitSet(path) {
						// Skip the file if the archive bit is not set
						return nil
					}
				}

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

		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Save metadata at the end
	saveMetadata(metadataFile, metadata)

	// Create the summary
	summary := "\n\n------------------------------------------\n"
	summary += fmt.Sprintf("New media files: %d\n", newMediaFilesCount)
	summary += fmt.Sprintf("Skipped media files: %d\n", skippedMediaFilesCount)
	summary += "------------------------------------------\n"
	summary += fmt.Sprintf("Copied %d other files\n", copiedFilesCount)
	summary += "Types:\n"
	for ext, count := range fileExtensionCount {
		summary += fmt.Sprintf("  %s: %d\n", ext, count)
	}

	// Add updated media files to the summary
	if len(updatedMediaFiles) > 0 {
		summary += "------------------------------------------\n"
		summary += "Updated media files:\n"
		for mediaFile, hash := range updatedMediaFiles {
			summary += fmt.Sprintf("  %s (new hash: %s)\n", mediaFile, hash)
		}
	}

	// Add timing information and regexp to the summary
	totalTime := time.Since(startTime)
	averageMediaFileTime := mediaFileProcessingTime.Seconds() / float64(newMediaFilesCount+len(updatedMediaFiles))
	summary += "------------------------------------------\n"
	summary += fmt.Sprintf("Total time taken: %.2f seconds\n", totalTime.Seconds())
	summary += fmt.Sprintf("Average time per new/updated media file: %.2f seconds\n", averageMediaFileTime)
	summary += fmt.Sprintf("Regular expression used: %s\n", pathRegexp)

	// Print the summary to standard out
	fmt.Print(summary)

	// Write the summary to a timestamped file
	timestamp := time.Now().Format("20060102-150405")
	summaryFile := filepath.Join(oneDriveFolder, fmt.Sprintf("sync-summary-%s.txt", timestamp))
	if err := os.WriteFile(summaryFile, []byte(summary), 0644); err != nil {
		fmt.Printf("Error writing summary file: %v\n", err)
	}
}

func handleFileEncryptionAndWriting(oneDriveFolder, relativePath, hashString string, data []byte, key []byte, info os.FileInfo, encryptFile bool) error {
	var destPath string
	var fileData []byte

	if encryptFile {
		// Create the destination path with hash
		destPath = filepath.Join(oneDriveFolder, relativePath+"-"+hashString+".encr")
		// Encrypt the file
		encryptedData, err := encrypt(data, key)
		if err != nil {
			return err
		}
		fileData = encryptedData
	} else {
		// Create the destination path without encryption
		destPath = filepath.Join(oneDriveFolder, relativePath)
		fileData = data
	}

	// Ensure the destination directory exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return err
	}

	// Write the file to the destination
	if err := os.WriteFile(destPath, fileData, info.Mode()); err != nil {
		return err
	}

	return nil
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
