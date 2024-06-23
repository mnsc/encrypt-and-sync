package main

import (
	"flag"
	"fmt"
	"os"
)

const keySize = 32

func main() {
	// Check if the encryption key is set and has the correct length
	key := os.Getenv("ENCRYPTION_KEY")
	if len(key) != keySize {
		fmt.Println("Error: ENCRYPTION_KEY environment variable must be set and 32 bytes long")
		os.Exit(1)
	}

	// Define command line options
	syncFlag := flag.Bool("sync", false, "Sync files to OneDrive")
	restorePath := flag.String("restore", "", "Destination path for restoring files")
	encryptFlag := flag.Bool("encrypt", false, "Encrypt files before copying")
	flag.Parse()

	if *syncFlag {
		syncFiles(*encryptFlag)
	} else if *restorePath != "" {
		restoreFiles(*restorePath)
	} else {
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("  -sync    Sync files to OneDrive")
	fmt.Println("  -restore Restore files from OneDrive")
	fmt.Println("  -encrypt Encrypt files before copying")
}
