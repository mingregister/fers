// main_improved.go
package main

import (
	"errors"
	"fmt"
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

func showFatalError(msg string) {
	a := app.New()
	w := a.NewWindow("启动失败")
	w.SetContent(widget.NewLabel(msg))
	w.Resize(fyne.NewSize(400, 200))
	dialog.ShowError(errors.New(msg), w)
	w.ShowAndRun() // 阻塞，用户关掉窗口后进程退出
}

func NewStorageClient(cfg *config.Storage) (storage.Client, error) {
	switch cfg.RemoteType {
	case "localhost":
		storageClient := storage.NewOSSMock(cfg.Localhost.Workdir)
		return storageClient, nil
	case "oss":
		storageClient, err := storage.NewOSSClient(
			cfg.Oss.Endpoint,
			cfg.Oss.AccessKeyID,
			cfg.Oss.AccessKeySecret,
			cfg.Oss.BucketName,
			cfg.Oss.Region,
			cfg.Oss.WorkDir,
		)
		return storageClient, err
	default:
		return nil, fmt.Errorf("unsupport storage %s", cfg.RemoteType)
	}
}

func main() {
	// Initialize configuration
	cfg, err := config.NewConfig()
	if err != nil {
		showFatalError(err.Error())
		return
	}

	// Create log widget first.
	// NOTE: logWidget需要先绑定到window才能使用.
	logWidget := widget.NewTextGrid()
	logWidget.Scroll = fyne.ScrollBoth
	logWidget.ShowWhitespace = true

	var f *os.File
	if cfg.Log != "" {
		var err error
		f, err = os.OpenFile(cfg.Log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			showFatalError(fmt.Sprintf("无法打开日志文件: %v", err))
			return
		}
		defer f.Close()
	}

	// Set up UI logger
	uiLogHandler := appui.NewUILogHandler(logWidget, &slog.HandlerOptions{
		Level:     slog.Level(cfg.LogLevel),
		AddSource: true,
	}, f)
	logger := slog.New(uiLogHandler)
	slog.SetDefault(logger)

	storageClient, err := NewStorageClient(&cfg.Storage)
	if err != nil {
		showFatalError(err.Error())
		return
	}

	cipherClient := crypto.NewAESGCM(cfg.CryptoKey)

	// Initialize file manager with UI logger
	fileManager := dir.NewFileManager(cfg, storageClient, logger, cipherClient)

	// Initialize UI with log widget
	ui := appui.NewAppUIWithLogWidget(fileManager, logger, logWidget)

	// Log startup message
	logger.Info("Application started successfully", slog.String("version", "1.0"))

	// Run UI
	ui.Run()
}
