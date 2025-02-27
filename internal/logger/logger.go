package logger

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Logger struct {
	db          *sql.DB
	environment string
	appName     string
	appVersion  string
	debug       bool
}

type Level string

const (
	GENERAL Level = "GENERAL"
	INFO    Level = "INFO"
	WARNING Level = "WARNING"
	SUCCESS Level = "SUCCESS"
	ERROR   Level = "ERROR"
)

// NewLogger creates a new logger instance
func NewLogger(db *sql.DB, environment, appName, appVersion string, debug bool) *Logger {
	return &Logger{
		db:          db,
		environment: environment,
		appName:     appName,
		appVersion:  appVersion,
		debug:       debug,
	}
}

// getCallerInfo gets the caller's file, function and line number
func getCallerInfo() (string, string, int) {
	var (
		module, function string
		line             int
	)

	// Skip 3 frames: getCallerInfo, log, and the actual logging function
	if pc, file, tmpLine, ok := runtime.Caller(3); ok {
		// Get full function name
		if fn := runtime.FuncForPC(pc); fn != nil {
			function = fn.Name()

			// Extract just the function name without package
			if idx := strings.LastIndex(function, "."); idx != -1 {
				function = function[idx+1:]
			}
		}

		// Get module path relative to project root
		if projectRoot := findProjectRoot(file); projectRoot != "" {
			if rel, err := filepath.Rel(projectRoot, file); err == nil {
				module = rel
			}
		} else {
			// Fallback to file name if project root not found
			module = filepath.Base(file)
		}

		line = tmpLine
	}

	return module, function, line
}

// findProjectRoot looks for the project root by searching for go.mod
func findProjectRoot(currentPath string) string {
	dir := filepath.Dir(currentPath)
	for dir != "/" {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}

// log handles the actual logging
func (l *Logger) log(level Level, message string, err error) {
	timestamp := time.Now()
	processID := fmt.Sprintf("%d", os.Getpid())
	module, function, line := getCallerInfo()

	// Console output
	if l.debug {
		consoleMsg := fmt.Sprintf("[%s][%s] - %s | Module: %s | Function: %s | Line: %d",
			timestamp.Format("2006-01-02 15:04:05"),
			level,
			message,
			module,
			function,
			line)
		if err != nil {
			consoleMsg += fmt.Sprintf(" | Error: %v", err)
		}
		fmt.Println(consoleMsg)
	} else {
		consoleMsg := fmt.Sprintf("[%s][%s] - %s",
			timestamp.Format("2006-01-02 15:04:05"),
			level,
			message)
		if err != nil {
			consoleMsg += fmt.Sprintf(" - Error: %v", err)
		}
		fmt.Println(consoleMsg)
	}

	// Database logging
	query := `
		INSERT INTO logs (
			timestamp, process_id, level, message, module, function, 
			line, error, environment, app_name, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	_, dbErr := l.db.Exec(
		query,
		timestamp,
		processID,
		level,
		message,
		module,
		function,
		line,
		errorMsg,
		l.environment,
		l.appName,
		timestamp,
	)

	if dbErr != nil {
		fmt.Printf("Failed to write to log database: %v\n", dbErr)
	}
}

// Info logs informational messages
func (l *Logger) Info(message string) {
	l.log(INFO, message, nil)
}

func (l *Logger) Success(message string) {
	l.log(SUCCESS, message, nil)
}

func (l *Logger) Warning(message string) {
	l.log(WARNING, message, nil)
}

// Error logs error information
func (l *Logger) Error(message string, err error) {
	l.log(ERROR, message, err)
}
