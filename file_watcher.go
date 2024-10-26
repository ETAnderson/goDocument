package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher struct
type FileWatcher struct {
	watcher *fsnotify.Watcher
	parser  *FileParser
}

// NewFileWatcher initializes a new FileWatcher
func NewFileWatcher(parser *FileParser) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FileWatcher{
		watcher: watcher,
		parser:  parser,
	}, nil
}

// Watch starts watching the specified directory
func (fw *FileWatcher) Watch(dir string) {
	// Add the directory to be watched
	err := fw.watcher.Add(dir)
	if err != nil {
		log.Fatalf("Failed to watch directory: %s", err)
	}

	// Handle signals for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case event, ok := <-fw.watcher.Events:
				if !ok {
					return
				}
				// Ignore remove events
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					continue
				}

				// Parse the changed file
				fmt.Printf("Event: %s\n", event)
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					fw.parser.ParseFile(event.Name)
				}
			case err, ok := <-fw.watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("Error: %s\n", err)
			case <-signalChan:
				fmt.Println("\nReceived interrupt signal, shutting down...")
				return
			}
		}
	}()
}
