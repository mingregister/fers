// main_improved.go
package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/mingregister/fers/pkg/appui"
	"github.com/mingregister/fers/pkg/config"
	"github.com/mingregister/fers/pkg/dir"
	"github.com/mingregister/fers/pkg/storage"
)

// initLogger initializes the logger
func initLogger(w io.Writer) *slog.Logger {
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})
	return slog.New(handler)
}

func main() {
	// Initialize configuration
	cfg := config.NewConfig()

	// Initialize logger
	logFile, err := os.OpenFile(cfg.Log, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}
	defer logFile.Close()

	logger := initLogger(logFile)
	slog.SetDefault(logger)

	// Initialize storage client
	storageClient := storage.NewOSSMock(cfg.OssDir)

	// Initialize file manager
	fileManager := dir.NewFileManager(cfg, storageClient, logger)

	// Initialize and run UI
	ui := appui.NewAppUI(fileManager, logger)
	ui.Run()
}
