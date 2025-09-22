// main_improved.go
package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/mingregister/fers/pkg/appui"
	"github.com/mingregister/fers/pkg/config"
	"github.com/mingregister/fers/pkg/crypto"
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

func showFatalError(msg string) {
	a := app.New()
	w := a.NewWindow("启动失败")
	w.SetContent(widget.NewLabel(msg))
	w.Resize(fyne.NewSize(400, 200))
	dialog.ShowError(errors.New(msg), w)
	w.ShowAndRun() // 阻塞，用户关掉窗口后进程退出
}

func main() {
	// Initialize configuration
	cfg, err := config.NewConfig()
	if err != nil {
		showFatalError(err.Error())
		return
	}

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

	cipherClient := crypto.NewAESGCM(cfg.CryptoKey)

	// Initialize file manager
	fileManager := dir.NewFileManager(cfg, storageClient, logger, cipherClient)

	// Initialize and run UI
	ui := appui.NewAppUI(fileManager, logger)
	ui.Run()
}
