package appui

import (
	"context"
	"fmt"
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
	"github.com/mingregister/fers/pkg/dir"
)

// AppUI manages the user interface
type AppUI struct {
	app         fyne.App
	window      fyne.Window
	fileManager *dir.FileManager
	logger      *slog.Logger

	// UI components
	list          *widget.List
	items         []string
	selectedIndex int
	selectedName  string

	// Directory navigation
	currentDir string // 当前显示的目录
	dirLabel   *widget.Label

	// Operation management
	operationMutex sync.Mutex
	cancelFunc     context.CancelFunc
}

// NewAppUI creates a new AppUI instance
func NewAppUI(fileManager *dir.FileManager, logger *slog.Logger) *AppUI {
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
		currentDir:    fileManager.GetWorkingDir(), // 初始化为workingDir
	}

	ui.setupUI()
	return ui
}

// setupUI initializes the user interface
func (ui *AppUI) setupUI() {
	// Directory labels
	workingDirLabel := widget.NewLabel("Working dir: " + ui.fileManager.GetWorkingDir())
	ui.dirLabel = widget.NewLabel("Current dir: " + ui.currentDir)

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

	// Navigation buttons
	navButtons := container.NewHBox(
		widget.NewButton("Up", ui.goUpDirectory),
		widget.NewButton("Enter", ui.enterSelectedDirectory),
	)

	// Operation buttons
	buttons := container.NewVBox(
		navButtons,
		widget.NewSeparator(),
		ui.createEncryptUploadButton(),
		ui.createSyncDownloadButton(),
		ui.createSyncUploadButton(),
		widget.NewButton("Refresh", ui.refreshList),
		ui.createCancelButton(),
	)

	// Layout
	dirLabels := container.NewVBox(workingDirLabel, ui.dirLabel)
	ListPane := container.NewBorder(dirLabels, nil, nil, nil, ui.list)
	content := container.NewBorder(nil, nil, buttons, nil, ListPane)
	ui.window.SetContent(content)
}

// refreshItems updates the items list
func (ui *AppUI) refreshItems() {
	ui.items = dir.List(ui.currentDir)
}

// refreshList refreshes the UI list
func (ui *AppUI) refreshList() {
	ui.refreshItems()
	ui.list.Length = func() int { return len(ui.items) }
	ui.list.Refresh()
}

// goUpDirectory navigates to the parent directory
func (ui *AppUI) goUpDirectory() {
	// 清理当前路径
	cleanCurrentDir := filepath.Clean(ui.currentDir)
	cleanWorkingDir := filepath.Clean(ui.fileManager.GetWorkingDir())

	// 不能超出workingDir的范围
	if cleanCurrentDir == cleanWorkingDir {
		dialog.ShowInformation("Info", "Already at working directory root", ui.window)
		return
	}

	parentDir := filepath.Dir(cleanCurrentDir)

	// 使用相对路径检查是否在workingDir范围内
	relPath, err := filepath.Rel(cleanWorkingDir, parentDir)
	if err != nil || strings.HasPrefix(relPath, "..") {
		parentDir = cleanWorkingDir
	}

	ui.currentDir = parentDir
	ui.dirLabel.SetText("Current dir: " + ui.currentDir)
	ui.refreshList()
	ui.selectedIndex = -1
	ui.selectedName = ""
}

// enterSelectedDirectory enters the selected directory
func (ui *AppUI) enterSelectedDirectory() {
	if ui.selectedIndex < 0 || ui.selectedIndex >= len(ui.items) {
		dialog.ShowInformation("Info", "Please select a directory first", ui.window)
		return
	}

	ui.enterDirectory(ui.selectedName)
}

// enterDirectory enters the specified directory
func (ui *AppUI) enterDirectory(dirName string) {
	fullPath := filepath.Join(ui.currentDir, dirName)

	// 检查是否是目录
	info, err := os.Stat(fullPath)
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to access %s: %w", dirName, err), ui.window)
		return
	}

	if !info.IsDir() {
		dialog.ShowInformation("Info", "Selected item is not a directory", ui.window)
		return
	}

	// 清理路径并确保不会超出workingDir的范围
	cleanFullPath := filepath.Clean(fullPath)
	cleanWorkingDir := filepath.Clean(ui.fileManager.GetWorkingDir())

	// 使用相对路径检查是否在workingDir范围内
	relPath, err := filepath.Rel(cleanWorkingDir, cleanFullPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		dialog.ShowInformation("Info", "Cannot navigate outside working directory", ui.window)
		return
	}

	ui.currentDir = cleanFullPath
	ui.dirLabel.SetText("Current dir: " + ui.currentDir)
	ui.refreshList()
	ui.selectedIndex = -1
	ui.selectedName = ""
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
			// 使用当前目录的完整路径
			fullPath := filepath.Join(ui.currentDir, name)

			info, err := os.Stat(fullPath)
			if err != nil {
				return fmt.Errorf("failed to stat file %s: %w", fullPath, err)
			}

			// 计算相对于workingDir的路径
			relativePath, err := filepath.Rel(ui.fileManager.GetWorkingDir(), fullPath)
			if err != nil {
				return fmt.Errorf("failed to get relative path for %s: %w", fullPath, err)
			}

			if info.IsDir() {
				return ui.fileManager.EncryptAndUploadDirectory(ctx, fullPath)
			} else {
				return ui.fileManager.EncryptAndUploadFile(fullPath, relativePath)
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
