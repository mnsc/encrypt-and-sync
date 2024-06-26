package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func restoreFiles(oneDriveFolder, destinationPath string, key []byte, test bool) {
	metadataFile := filepath.Join(oneDriveFolder, "metadata.json")

	metadata := loadMetadata(metadataFile)
	latestFiles := getLatestFiles(metadata)

	if test {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		keys := make([]string, 0, len(latestFiles))
		for k := range latestFiles {
			keys = append(keys, k)
		}
		randomKey := keys[r.Intn(len(keys))]
		latestFiles = map[string]FileMetadata{randomKey: latestFiles[randomKey]}
	}

	// Restore files tracked in metadata
	for _, file := range latestFiles {
		// Construct the source path
		sourcePath := filepath.Join(oneDriveFolder, file.OriginalPath+"-"+file.Hash+".encr")

		// Read the encrypted file
		encryptedData, err := os.ReadFile(sourcePath)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", sourcePath, err)
			continue
		}

		// Decrypt the file
		data, err := decrypt(encryptedData, key)
		if err != nil {
			fmt.Printf("Error decrypting file %s: %v\n", sourcePath, err)
			continue
		}

		// Construct the destination path
		destPath := filepath.Join(destinationPath, file.OriginalPath)

		// Ensure the destination directory exists
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
			fmt.Printf("Error creating directory %s: %v\n", destDir, err)
			continue
		}

		// Write the decrypted file to the destination
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			fmt.Printf("Error writing file %s: %v\n", destPath, err)
		} else {
			fmt.Printf("Restored file: %s\n", destPath) // Print restored file
		}
	}

	if test {
		fmt.Printf("Will not restore other files in -test\n")
	} else {
		// Restore other files not tracked in metadata
		err := filepath.Walk(oneDriveFolder, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() && !strings.HasSuffix(path, ".encr") {
				// Construct the destination path
				relativePath, err := filepath.Rel(oneDriveFolder, path)
				if err != nil {
					return err
				}
				destPath := filepath.Join(destinationPath, relativePath)

				// Ensure the destination directory exists
				destDir := filepath.Dir(destPath)
				if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
					return err
				}

				// Copy the file to the destination
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if err := os.WriteFile(destPath, data, info.Mode()); err != nil {
					return err
				} else {
					fmt.Printf("Restored file: %s\n", destPath) // Print restored file
				}
			}

			return nil
		})

		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	fmt.Println("Restore completed.")
}

func getLatestFiles(metadata []FileMetadata) map[string]FileMetadata {
	latestFiles := make(map[string]FileMetadata)
	for _, file := range metadata {
		if existingFile, exists := latestFiles[file.OriginalPath]; !exists || file.ModTime > existingFile.ModTime {
			latestFiles[file.OriginalPath] = file
		}
	}
	return latestFiles
}
