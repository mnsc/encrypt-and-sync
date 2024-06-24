package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

const keySize = 32

func main() {
	// Check if the encryption key is set and has the correct length
	keyString := os.Getenv("ENCRYPTION_KEY")
	var key []byte
	if len(keyString) != keySize {
		fmt.Println("ENCRYPTION_KEY environment variable not set or incorrect length. Please enter the encryption key:")
		reader := bufio.NewReader(os.Stdin)
		keyString, _ = reader.ReadString('\n')
		keyString = strings.TrimSpace(keyString)
		if len(keyString) != keySize {
			fmt.Println("Error: ENCRYPTION_KEY must be 32 bytes long")
			os.Exit(1)
		}
		key = []byte(keyString)
	}

	// Define command line options
	syncFlag := flag.Bool("sync", false, "Sync files to OneDrive")
	restorePath := flag.String("restore", "", "Destination path for restoring files")
	encryptFlag := flag.Bool("encrypt", true, "Encrypt files before copying")
	sourceFolder := flag.String("source", "", "Source folder for syncing files")
	oneDriveFolder := flag.String("onedrive", "", "OneDrive folder for syncing files")
	pathRegexp := flag.String("pathregexp", ".*", "Regular expression to match file paths for processing")
	testFlag := flag.Bool("test", false, "Restore one random photo from metadata")
	metadataFlag := flag.Bool("metadata", false, "Print information about the metadata file")
	flag.Parse()

	if *syncFlag {
		if *sourceFolder == "" || *oneDriveFolder == "" {
			fmt.Println("Error: -source and -onedrive flags must be specified for syncing files")
			os.Exit(1)
		}
		syncFiles(*sourceFolder, *oneDriveFolder, *encryptFlag, key, *pathRegexp)
	} else if *restorePath != "" {
		restoreFiles(*oneDriveFolder, *restorePath, key, *testFlag)
	} else if *metadataFlag {
		printMetadataInfo(*oneDriveFolder) // Call the new function
	} else {
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Usage:")
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("  -%s\t%s\n", f.Name, f.Usage)
	})
}
