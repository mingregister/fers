// main_improved.go
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/mingregister/fers/pkg/config"
	"github.com/mingregister/fers/pkg/crypto"
	"github.com/mingregister/fers/pkg/dir"
	"github.com/mingregister/fers/pkg/storage"
)

const (
	defaultFileMode = 0o644
	defaultDirMode  = 0o755
)

// FileManager handles file operations with encryption and remote storage
type FileManager struct {
	config     *config.Config
	storage    storage.Client
	workingDir string
	cryptoKey  []byte
	logger     *slog.Logger
}

// NewFileManager creates a new FileManager instance
func NewFileManager(cfg *config.Config, storage storage.Client, logger *slog.Logger) *FileManager {
	return &FileManager{
		config:     cfg,
		storage:    storage,
		workingDir: cfg.TargetDir,
		cryptoKey:  crypto.DeriveKeyFromPassword(cfg.CryptoKey),
		logger:     logger,
	}
}

// EncryptAndUploadFile encrypts and uploads a single file
func (fm *FileManager) EncryptAndUploadFile(filePath, relativePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	encrypted, err := crypto.EncryptAESGCM(fm.cryptoKey, data)
	if err != nil {
		return fmt.Errorf("failed to encrypt file %s: %w", filePath, err)
	}

	if err := fm.storage.Upload(filepath.ToSlash(relativePath), encrypted); err != nil {
		return fmt.Errorf("failed to upload file %s: %w", relativePath, err)
	}

	fm.logger.Info("File uploaded successfully", slog.String("path", relativePath))
	return nil
}

// EncryptAndUploadDirectory recursively encrypts and uploads a directory
func (fm *FileManager) EncryptAndUploadDirectory(ctx context.Context, dirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return fmt.Errorf("walk error at %s: %w", path, err)
		}

		if info.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(fm.workingDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		return fm.EncryptAndUploadFile(path, relativePath)
	})
}

// DownloadAndDecryptFile downloads and decrypts a single file
func (fm *FileManager) DownloadAndDecryptFile(remotePath, localPath string) error {
	encrypted, err := fm.storage.Download(remotePath)
	if err != nil {
		return fmt.Errorf("failed to download file %s: %w", remotePath, err)
	}

	decrypted, err := crypto.DecryptAESGCM(fm.cryptoKey, encrypted)
	if err != nil {
		return fmt.Errorf("failed to decrypt file %s: %w", remotePath, err)
	}

	dir := filepath.Dir(localPath)
	if err := os.MkdirAll(dir, defaultDirMode); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(localPath, decrypted, defaultFileMode); err != nil {
		return fmt.Errorf("failed to write file %s: %w", localPath, err)
	}

	fm.logger.Info("File downloaded and decrypted successfully", slog.String("path", localPath))
	return nil
}

// SyncDownload downloads missing files from remote storage
func (fm *FileManager) SyncDownload(ctx context.Context) error {
	remoteFiles, err := fm.storage.List("")
	if err != nil {
		return fmt.Errorf("failed to list remote files: %w", err)
	}

	localFiles := dir.List(fm.workingDir)
	localSet := make(map[string]bool, len(localFiles))
	for _, file := range localFiles {
		localSet[file] = true
	}

	for _, remotePath := range remoteFiles {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		parts := strings.Split(remotePath, "/")
		topLevel := parts[0]

		if !localSet[topLevel] {
			localPath := filepath.Join(fm.workingDir, remotePath)
			if err := fm.DownloadAndDecryptFile(remotePath, localPath); err != nil {
				fm.logger.Error("Failed to download file", slog.String("path", remotePath), slog.String("error", err.Error()))
				continue
			}
		}
	}

	return nil
}

// SyncUpload uploads missing local files to remote storage
func (fm *FileManager) SyncUpload(ctx context.Context) error {
	remoteFiles, err := fm.storage.List("")
	if err != nil {
		return fmt.Errorf("failed to list remote files: %w", err)
	}

	remoteSet := make(map[string]bool, len(remoteFiles))
	for _, file := range remoteFiles {
		remoteSet[file] = true
		remoteSet[strings.Split(file, "/")[0]] = true
	}

	return filepath.Walk(fm.workingDir, func(path string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return fmt.Errorf("walk error at %s: %w", path, err)
		}

		if info.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(fm.workingDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		relativeSlash := filepath.ToSlash(relativePath)
		if !remoteSet[relativeSlash] {
			if err := fm.EncryptAndUploadFile(path, relativePath); err != nil {
				fm.logger.Error("Failed to upload file", slog.String("path", relativePath), slog.String("error", err.Error()))
			}
		}

		return nil
	})
}

// AppUI manages the user interface
type AppUI struct {
	app         fyne.App
	window      fyne.Window
	fileManager *FileManager
	logger      *slog.Logger

	// UI components
	list          *widget.List
	items         []string
	selectedIndex int
	selectedName  string

	// Operation management
	operationMutex sync.Mutex
	cancelFunc     context.CancelFunc
}

// NewAppUI creates a new AppUI instance
func NewAppUI(fileManager *FileManager, logger *slog.Logger) *AppUI {
	app := app.New()
	window := app.NewWindow("File Encrypt & Remote Storage")
	window.Resize(fyne.NewSize(1000, 600))
	window.CenterOnScreen()

	ui := &AppUI{
		app:           app,
		window:        window,
		fileManager:   fileManager,
		logger:        logger,
		selectedIndex: -1,
	}

	ui.setupUI()
	return ui
}

// setupUI initializes the user interface
func (ui *AppUI) setupUI() {
	// Working directory label
	dirLabel := widget.NewLabel("Working dir: " + ui.fileManager.workingDir)

	// File list
	ui.refreshItems()
	ui.list = widget.NewList(
		func() int { return len(ui.items) },
		func() fyne.CanvasObject {
			return widget.NewLabel("template")
		},
		func(i int, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(ui.items[i])
		},
	)
	ui.list.OnSelected = func(i int) {
		ui.selectedIndex = i
		ui.selectedName = ui.items[i]
	}

	// Buttons
	buttons := container.NewVBox(
		ui.createEncryptUploadButton(),
		ui.createSyncDownloadButton(),
		ui.createSyncUploadButton(),
		widget.NewButton("Refresh", ui.refreshList),
		ui.createCancelButton(),
	)

	// Layout
	ListPane := container.NewBorder(dirLabel, nil, nil, nil, ui.list)
	content := container.NewBorder(nil, nil, buttons, nil, ListPane)
	ui.window.SetContent(content)
}

// refreshItems updates the items list
func (ui *AppUI) refreshItems() {
	ui.items = dir.List(ui.fileManager.workingDir)
}

// refreshList refreshes the UI list
func (ui *AppUI) refreshList() {
	ui.refreshItems()
	ui.list.Length = func() int { return len(ui.items) }
	ui.list.Refresh()
}

// createEncryptUploadButton creates the encrypt and upload button
func (ui *AppUI) createEncryptUploadButton() *widget.Button {
	return widget.NewButton("Encrypt & Upload", func() {
		if ui.selectedIndex < 0 || ui.selectedIndex >= len(ui.items) {
			dialog.ShowInformation("Info", "Please select a file or directory first", ui.window)
			return
		}

		ui.runOperation("Encrypt & Upload", func(ctx context.Context) error {
			name := ui.selectedName
			fullPath := filepath.Join(ui.fileManager.workingDir, name)

			info, err := os.Stat(fullPath)
			if err != nil {
				return fmt.Errorf("failed to stat file %s: %w", fullPath, err)
			}

			if info.IsDir() {
				return ui.fileManager.EncryptAndUploadDirectory(ctx, fullPath)
			} else {
				return ui.fileManager.EncryptAndUploadFile(fullPath, name)
			}
		})
	})
}

// createSyncDownloadButton creates the sync download button
func (ui *AppUI) createSyncDownloadButton() *widget.Button {
	return widget.NewButton("Sync Download", func() {
		ui.runOperation("Sync Download", func(ctx context.Context) error {
			err := ui.fileManager.SyncDownload(ctx)
			if err == nil {
				ui.refreshList()
			}
			return err
		})
	})
}

// createSyncUploadButton creates the sync upload button
func (ui *AppUI) createSyncUploadButton() *widget.Button {
	return widget.NewButton("Sync Upload", func() {
		ui.runOperation("Sync Upload", func(ctx context.Context) error {
			return ui.fileManager.SyncUpload(ctx)
		})
	})
}

// createCancelButton creates the cancel operation button
func (ui *AppUI) createCancelButton() *widget.Button {
	return widget.NewButton("Cancel Operation", func() {
		ui.operationMutex.Lock()
		defer ui.operationMutex.Unlock()

		if ui.cancelFunc != nil {
			ui.cancelFunc()
			ui.logger.Info("Operation cancelled by user")
		}
	})
}

// runOperation runs a long-running operation with proper error handling and cancellation
func (ui *AppUI) runOperation(operationName string, operation func(context.Context) error) {
	ui.operationMutex.Lock()
	defer ui.operationMutex.Unlock()

	// Cancel any existing operation
	if ui.cancelFunc != nil {
		ui.cancelFunc()
	}

	ctx, cancel := context.WithCancel(context.Background())
	ui.cancelFunc = cancel

	go func() {
		defer func() {
			ui.operationMutex.Lock()
			ui.cancelFunc = nil
			ui.operationMutex.Unlock()
		}()

		ui.logger.Info("Starting operation", slog.String("operation", operationName))

		if err := operation(ctx); err != nil {
			if err == context.Canceled {
				ui.logger.Info("Operation cancelled", slog.String("operation", operationName))
			} else {
				ui.logger.Error("Operation failed",
					slog.String("operation", operationName),
					slog.String("error", err.Error()))
				dialog.ShowError(err, ui.window)
			}
			return
		}

		ui.logger.Info("Operation completed successfully", slog.String("operation", operationName))
	}()
}

// Run starts the application
func (ui *AppUI) Run() {
	ui.window.ShowAndRun()
}

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
	logFile, err := os.OpenFile(cfg.Log, os.O_CREATE|os.O_RDWR|os.O_APPEND, defaultFileMode)
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
	fileManager := NewFileManager(cfg, storageClient, logger)

	// Initialize and run UI
	ui := NewAppUI(fileManager, logger)
	ui.Run()
}
