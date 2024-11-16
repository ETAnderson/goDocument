package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher struct
type FileWatcher struct {
	watcher    *fsnotify.Watcher
	parser     *FileParser
	logFile    *os.File
	recentLogs []string   // Store recent log entries
	mu         sync.Mutex // Mutex for thread-safe access to recentLogs
}

// NewFileWatcher initializes a new FileWatcher
func NewFileWatcher(parser *FileParser) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		return nil, err
	}

	// Create a date-stamped log file
	timestamp := time.Now().Format("2006-01-02") // Format: YYYY-MM-DD
	logFileName := fmt.Sprintf("logs/file_watcher_logs_%s.txt", timestamp)

	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		watcher:    watcher,
		parser:     parser,
		logFile:    logFile,
		recentLogs: []string{}, // Initialize the array for recent logs
	}, nil
}

// Watch starts watching the specified directory and its subdirectories
func (fw *FileWatcher) Watch(dir string) {
	// Add the directory to be watched
	if err := fw.watcher.Add(dir); err != nil {
		log.Fatalf("Failed to watch directory: %s", err)
	}

	// Watch all subdirectories recursively
	fw.watchSubdirectories(dir)

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
				log.Printf("Error: %s\n", err)
			case <-signalChan:
				fmt.Println("\nReceived interrupt signal, shutting down...")
				fw.Close() // Close resources before exit
				return
			}
		}
	}()
}

// watchSubdirectories recursively watches all subdirectories
func (fw *FileWatcher) watchSubdirectories(dir string) {
	// Use Walk to add all subdirectories to the watcher
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if err := fw.watcher.Add(path); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Error watching subdirectories: %v", err)
	}
}

// logEvent logs the file change event to the log file
func (fw *FileWatcher) logEvent(event fsnotify.Event) {
	fw.mu.Lock()
	defer fw.mu.Unlock() // Ensure to unlock the mutex after logging

	timestamp := time.Now().Format("15:04:05")                             // Format: HH:MM:SS
	logEntry := fmt.Sprintf("%s, %s, %s", timestamp, event.Name, event.Op) // Log format: HH:MM:SS, filename, operation

	// Check if the last logged entry is the same as the current log entry
	if len(fw.recentLogs) > 0 && fw.recentLogs[len(fw.recentLogs)-1] == logEntry {
		fw.recentLogs = []string{} // Clear recent logs if the current entry is a duplicate
		return                     // Skip logging if it's a duplicate of the last entry
	}

	// Write log entry to the file
	if _, err := fw.logFile.WriteString(logEntry + "\n"); err != nil {
		log.Printf("Error writing to log file: %s\n", err)
		return
	}

	// Append the log entry to recent logs
	fw.recentLogs = append(fw.recentLogs, logEntry)
}

// Wait blocks until the watcher is closed
func (fw *FileWatcher) Wait() {
	// Wait for a signal to close the watcher
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	<-signalChan // Block until we receive a signal
	fw.Close()   // Close the file watcher and log file
}

// Close closes the file watcher and log file
func (fw *FileWatcher) Close() {
	fw.watcher.Close()
	fw.logFile.Close()
}
