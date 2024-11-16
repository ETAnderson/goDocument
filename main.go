package main

import (
	"log"
	"os"
)

func main() {
	// Check if a directory path is provided as an argument
	if len(os.Args) < 2 {
		log.Fatal("Please provide a directory path to watch.")
	}

	dir := os.Args[1]

	// Create a new FileParser
	parser := NewFileParser()

	// Build the initial directory structure for references
	if err := BuildFileStructure(dir); err != nil {
		log.Fatalf("Error creating directory structure: %v", err)
	}

	// Create a new FileWatcher
	fileWatcher, err := NewFileWatcher(parser)
	if err != nil {
		log.Fatalf("Error initializing file watcher: %v", err)
	}

	// Start watching the specified directory
	fileWatcher.Watch(dir)

	// Wait for shutdown signal
	fileWatcher.Wait()

	log.Println("File watcher stopped.")
}
