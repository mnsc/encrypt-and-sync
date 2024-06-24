package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type FileMetadata struct {
	OriginalPath string `json:"original_path"`
	Hash         string `json:"hash"`
	ModTime      int64  `json:"mod_time"`
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

func printMetadataInfo(oneDriveFolder string) {

	metadataFile := filepath.Join(oneDriveFolder, "metadata.json")
	metadata := loadMetadata(metadataFile)

	fmt.Println("Metadata information:")
	fmt.Printf("Metadata file: %s\n", metadataFile)
	fmt.Printf("Metadata length: %d\n", len(metadata))

	// Count occurrences of each file path
	fileCount := make(map[string]int)
	for _, entry := range metadata {
		fileCount[entry.OriginalPath]++
	}

	// Print details of files that appear more than once
	for path, count := range fileCount {
		if count > 1 {
			fmt.Printf("\nFile: %s\n", path)
			for _, entry := range metadata {
				if entry.OriginalPath == path {
					fmt.Printf(" %s...%s (%s)\n", entry.Hash[:4], entry.Hash[len(entry.Hash)-4:], time.Unix(entry.ModTime, 0).Format(time.RFC3339))
				}
			}
		}
	}
}
