package main

import (
	"os"
	"path/filepath"
)

// buildFileStructure creates a corresponding directory structure in the "references" folder
func buildFileStructure(watchedDir string) error {
	// Define the references directory
	referenceDir := "references"

	// Walk through the watched directory
	return filepath.Walk(watchedDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Construct the corresponding path in the references directory
		relPath, _ := filepath.Rel(watchedDir, path)
		destPath := filepath.Join(referenceDir, relPath)

		if info.IsDir() {
			// Create the directory in the references structure
			return os.MkdirAll(destPath, os.ModePerm)
		}
		return nil // Return nil for files; we only need to create directories
	})
}
