package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher struct
type FileWatcher struct {
	watcher    *fsnotify.Watcher
	parser     *FileParser
	logFile    *os.File
	recentLogs map[string]struct{} // Store recent log entries to prevent duplicates
}

// NewFileWatcher initializes a new FileWatcher
func NewFileWatcher(parser *FileParser) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Create a date-stamped log file
	timestamp := time.Now().Format("2006-01-02") // Format: YYYY-MM-DD
	logFileName := fmt.Sprintf("logs/file_watcher_logs_%s.txt", timestamp)

	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		watcher:    watcher,
		parser:     parser,
		logFile:    logFile,
		recentLogs: make(map[string]struct{}), // Initialize the map
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

				// Log the event
				fw.logEvent(event)

				// Parse the changed file
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

// logEvent logs the file change event to the log file
func (fw *FileWatcher) logEvent(event fsnotify.Event) {
	// Create a unique log key
	logKey := fmt.Sprintf("%s:%s", event.Name, event.Op)

	// Check for duplicates
	if _, exists := fw.recentLogs[logKey]; exists {
		return // Skip logging if it's a duplicate
	}
	fw.recentLogs[logKey] = struct{}{} // Mark this log entry as seen

	timestamp := time.Now().Format("15:04:05") // Format: HH:MM:SS
	logEntry := fmt.Sprintf("%s, %s, %s\n", event.Name, event.Op, timestamp)
	if _, err := fw.logFile.WriteString(logEntry); err != nil {
		fmt.Printf("Error writing to log file: %s\n", err)
	}
}

// Close closes the file watcher and log file
func (fw *FileWatcher) Close() {
	fw.watcher.Close()
	fw.logFile.Close()
}
