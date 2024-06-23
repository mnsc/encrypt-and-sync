package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileMetadata struct {
	OriginalPath string `json:"original_path"`
	Hash         string `json:"hash"`
	ModTime      int64  `json:"mod_time"`
}

func main() {
	rootFolder := "C:\\Temp\\SourceTest"
	oneDriveFolder := "C:\\Temp\\OneDriveTest"
	metadataFile := filepath.Join(oneDriveFolder, "metadata.json")

	metadata := loadMetadata(metadataFile)
	metadataMap := createMetadataMap(metadata)

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

				// Check if the file with the same path and hash already exists in metadata
				if !isFileInMetadataMap(metadataMap, relativePath, hashString) {
					// Create the destination path with hash
					destPath = filepath.Join(oneDriveFolder, relativePath+"-"+hashString+".encr")

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

					// Update metadata
					newMetadata := FileMetadata{
						OriginalPath: relativePath,
						Hash:         hashString,
						ModTime:      modTime,
					}
					metadata = append(metadata, newMetadata)
					metadataMap[relativePath+"-"+hashString] = newMetadata
				} else {
					fmt.Printf("File already exists in metadata, skipping: %s\n", relativePath)
				}
			} else {
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
				fmt.Printf("Copied: %s -> %s\n", path, destPath)
			}

		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Save metadata at the end
	saveMetadata(metadataFile, metadata)
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
		key := entry.OriginalPath + "-" + entry.Hash
		metadataMap[key] = entry
	}
	return metadataMap
}

func isFileInMetadataMap(metadataMap map[string]FileMetadata, path, hash string) bool {
	key := path + "-" + hash
	_, exists := metadataMap[key]
	return exists
}
