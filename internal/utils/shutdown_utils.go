package utils

import (
	"context"
	"database/sql"
	"dbaTools/internal/logger"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// ShutdownHandler manages graceful application shutdown
type ShutdownHandler struct {
	db           *sql.DB
	logger       *logger.Logger
	cleanupFuncs []func() error
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewShutdownHandler creates a new shutdown handler instance
func NewShutdownHandler(db *sql.DB, logger *logger.Logger) *ShutdownHandler {
	ctx, cancel := context.WithCancel(context.Background())
	return &ShutdownHandler{
		db:           db,
		logger:       logger,
		cleanupFuncs: make([]func() error, 0),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// RegisterCleanupFunc registers a function to be called during shutdown
func (sh *ShutdownHandler) RegisterCleanupFunc(cleanup func() error) {
	sh.cleanupFuncs = append(sh.cleanupFuncs, cleanup)
}

// GetContext returns the shutdown context
func (sh *ShutdownHandler) GetContext() context.Context {
	return sh.ctx
}

// WaitForShutdown waits for shutdown signals and performs cleanup
func (sh *ShutdownHandler) WaitForShutdown() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	sig := <-signalChan
	sh.logger.Info(fmt.Sprintf("Received shutdown signal: %v", sig))

	// Cancel context to notify all operations
	sh.cancel()

	// Create timeout context for cleanup operations
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute all cleanup functions
	for _, cleanup := range sh.cleanupFuncs {
		sh.wg.Add(1)
		go func(cleanupFunc func() error) {
			defer sh.wg.Done()
			if err := cleanupFunc(); err != nil {
				sh.logger.Error("Cleanup operation failed", err)
			}
		}(cleanup)
	}

	// Wait for cleanup with timeout
	doneChan := make(chan struct{})
	go func() {
		sh.wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		sh.logger.Success("Graceful shutdown completed")
	case <-timeoutCtx.Done():
		sh.logger.Warning("Shutdown timed out, forcing exit")
	}

	// Final database cleanup
	if sh.db != nil {
		if err := sh.db.Close(); err != nil {
			sh.logger.Error("Error closing database connection", err)
		}
	}

	os.Exit(0)
}

// InitiateShutdown triggers a manual shutdown
func (sh *ShutdownHandler) InitiateShutdown() {
	sh.logger.Info("Manual shutdown initiated")
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		sh.logger.Error("Failed to find process", err)
		return
	}
	process.Signal(syscall.SIGTERM)
}
