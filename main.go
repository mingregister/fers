// main_improved.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

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

// startPprofServer starts the pprof HTTP server if enabled
func startPprofServer(cfg *config.Pprof, logger *slog.Logger) *http.Server {
	if !cfg.Enabled {
		return nil
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Starting pprof server",
			slog.String("address", addr),
			slog.String("endpoints", "http://"+addr+"/debug/pprof/"))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Pprof server failed", slog.String("error", err.Error()))
		}
	}()

	return server
}

// gracefulShutdown handles graceful shutdown of the pprof server
func gracefulShutdown(server *http.Server, logger *slog.Logger) {
	if server == nil {
		return
	}

	// Create a channel to listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down pprof server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Pprof server shutdown failed", slog.String("error", err.Error()))
		} else {
			logger.Info("Pprof server shutdown completed")
		}
	}()
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

	logOpt := &slog.HandlerOptions{
		Level:     slog.Level(cfg.LogLevel),
		AddSource: true,
	}

	f, err := os.OpenFile(cfg.Log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		showFatalError(fmt.Sprintf("无法打开日志文件: %v", err))
		return
	}
	defer f.Close()

	th := slog.NewTextHandler(f, logOpt) // Validate options
	textLogger := slog.New(th)

	// Set up UI logger
	uiLogHandler := appui.NewUILogHandler(logWidget, logOpt, textLogger)
	logger := slog.New(uiLogHandler)
	slog.SetDefault(logger)

	// Start pprof server if enabled
	pprofServer := startPprofServer(&cfg.Pprof, textLogger)
	if pprofServer != nil {
		gracefulShutdown(pprofServer, textLogger)
	}

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
