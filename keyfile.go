package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type KeyFileContent struct {
	KnownProperty  string `json:"known_property"`
	RandomProperty string `json:"random_property"`
}

// handleKeyFile manages the key file in the specified OneDrive folder.
// If a keyfile doesn't exist, it creates a new one and encrypts it with the provided key.
// If a keyfile does exist, it decrypts it with the provided key and verifies that the json contains a known property.
// This is to ensure that the same cryptographic key is used between syncs.
func handleKeyFile(oneDriveFolder string, key []byte) error {
	keyFilePath := filepath.Join(oneDriveFolder, "keyfile.encr")

	// Check if the key file exists
	if _, err := os.Stat(keyFilePath); os.IsNotExist(err) {
		// Create a new key file
		content := KeyFileContent{
			KnownProperty:  "known_value",
			RandomProperty: generateRandomString(2048),
		}
		data, err := json.Marshal(content)
		if err != nil {
			return err
		}
		encryptedData, err := encrypt(data, key)
		if err != nil {
			return err
		}
		if err := os.WriteFile(keyFilePath, encryptedData, 0644); err != nil {
			return err
		}
	} else {
		// Read and decrypt the existing key file
		encryptedData, err := os.ReadFile(keyFilePath)
		if err != nil {
			return err
		}
		data, err := decrypt(encryptedData, key)
		if err != nil {
			return err
		}
		var content KeyFileContent
		if err := json.Unmarshal(data, &content); err != nil {
			return err
		}
		// Verify the known property
		if content.KnownProperty != "known_value" {
			return errors.New("key file verification failed")
		}
	}
	return nil
}

func generateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	for i := range bytes {
		bytes[i] = letters[bytes[i]%byte(len(letters))]
	}
	return string(bytes)
}
