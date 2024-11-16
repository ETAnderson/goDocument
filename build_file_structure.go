package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// BuildFileStructure replicates srcDir's structure inside the "references" directory
// and generates JSON files for .go files found.
func BuildFileStructure(srcDir string) error {
	referencesDir := "references"

	// Walk through the source directory
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %q: %v\n", path, err)
			return err
		}

		// Skip the references directory and its subdirectories
		if strings.HasPrefix(path, referencesDir) {
			return nil
		}

		// Calculate the relative path from srcDir
		relativePath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Skip processing the root directory itself
		if relativePath == "." {
			return nil
		}

		// Construct the target path under references/
		targetPath := filepath.Join(referencesDir, relativePath)

		if info.IsDir() {
			// Ensure directories are replicated in references
			if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
				log.Printf("Error creating directory %q: %v\n", targetPath, err)
				return err
			}
		} else if filepath.Ext(path) == ".go" {
			// Process only .go files
			parser := NewFileParser()
			parser.ParseFile(path)

			// Convert targetPath from .go to .json
			jsonFilePath := targetPath[:len(targetPath)-len(filepath.Ext(targetPath))] + ".json"

			// Write JSON strictly to the references directory
			if err := parser.writeJSONToFile(filepath.Join(referencesDir, jsonFilePath)); err != nil {
				log.Printf("Error writing JSON to file: %v\n", err)
				return err
			}
		}

		return nil
	})
}
