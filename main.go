// main_improved.go
package main

import (
	"errors"
	"log/slog"

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

	// Set up UI logger
	uiLogHandler := appui.NewUILogHandler(logWidget, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	})
	logger := slog.New(uiLogHandler)
	slog.SetDefault(logger)

	storageClient := storage.NewOSSMock(cfg.StorageMockDir)
	// storageClient, _ := storage.NewOSSClient(
	// 	cfg.OSS.Endpoint,
	// 	cfg.OSS.AccessKeyID,
	// 	cfg.OSS.AccessKeySecret,
	// 	cfg.OSS.BucketName,
	// 	cfg.OSS.Region,
	// 	cfg.OSS.WorkDir,
	// )

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
