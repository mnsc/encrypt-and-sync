package main

import (
	"encoding/json"
	"os"
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
