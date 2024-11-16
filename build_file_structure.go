package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// BuildFileStructure replicates the srcDir structure inside the "references" directory
// and generates JSON files for all .go files found.
func BuildFileStructure(srcDir string) error {
	referencesDir := "references"
	baseDirName := filepath.Base(srcDir)

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing path %q: %v\n", path, err)
			return err
		}

		// Skip unintended nested references directory
		relativePath, err := filepath.Rel(srcDir, path)
		if err != nil {
			log.Printf("Error calculating relative path: %v\n", err)
			return err
		}
		if strings.Contains(relativePath, referencesDir) {
			log.Printf("Skipping unintended nested references directory: %q\n", relativePath)
			return nil
		}

		// Create the target path under references
		targetPath := filepath.Join(referencesDir, baseDirName, relativePath)

		if info.IsDir() {
			// Create directories
			if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
				log.Printf("Error creating directory %q: %v\n", targetPath, err)
				return err
			}
		} else if filepath.Ext(path) == ".go" {
			// Process .go files
			parser := NewFileParser()
			parser.ParseFile(path)

			// Define JSON file path
			jsonFilePath := strings.TrimSuffix(targetPath, filepath.Ext(targetPath)) + ".json"

			if err := parser.writeJSONToFile(jsonFilePath); err != nil {
				log.Printf("Error writing JSON to file %q: %v\n", jsonFilePath, err)
				return err
			}
		}

		return nil
	})
}
