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
	done       chan struct{}       // Channel to signal completion
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
		done:       make(chan struct{}),       // Initialize the done channel
	}, nil
}

// Watch starts watching the specified directory
func (fw *FileWatcher) Watch(dir string) {
	// Add the directory to be watched
	if err := fw.watcher.Add(dir); err != nil {
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
				fw.debounceParsing(event)

			case err, ok := <-fw.watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("Error: %s\n", err)

			case <-signalChan:
				fmt.Println("\nReceived interrupt signal, shutting down...")
				fw.Close()
				return
			}
		}
	}()
}

// debounceParsing implements debouncing for file parsing
func (fw *FileWatcher) debounceParsing(event fsnotify.Event) {
	logKey := fmt.Sprintf("%s:%s", event.Name, event.Op)

	// Check for duplicates
	if _, exists := fw.recentLogs[logKey]; exists {
		return // Skip logging if it's a duplicate
	}
	fw.recentLogs[logKey] = struct{}{} // Mark this log entry as seen

	// Wait for a short period before parsing to prevent rapid triggers
	time.AfterFunc(100*time.Millisecond, func() {
		fw.parser.ParseFile(event.Name)
		fw.logEvent(event)
	})
}

// logEvent logs the file change event to the log file
func (fw *FileWatcher) logEvent(event fsnotify.Event) {
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
	close(fw.done)
}

// Wait blocks until the file watcher is done
func (fw *FileWatcher) Wait() {
	<-fw.done
}
