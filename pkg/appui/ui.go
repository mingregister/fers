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
		ui.createDownloadSpecificButton(),
		ui.createSyncUploadButton(),
		ui.createDeleteLocalFileButton(),
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
	// 清除选择状态
	ui.list.UnselectAll()
	ui.selectedIndex = -1
	ui.selectedName = ""
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
		// 检查是否有选中的项目
		if ui.selectedIndex < 0 || ui.selectedIndex >= len(ui.items) || ui.selectedName == "" {
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

// createDownloadSpecificButton creates the download specific file button
func (ui *AppUI) createDownloadSpecificButton() *widget.Button {
	return widget.NewButton("Download Specific", func() {
		ui.showRemoteFileDialog()
	})
}

// createDeleteLocalFileButton creates the delete local file button
func (ui *AppUI) createDeleteLocalFileButton() *widget.Button {
	return widget.NewButton("Delete Local File", func() {
		// 检查是否有选中的项目
		if ui.selectedIndex < 0 || ui.selectedIndex >= len(ui.items) || ui.selectedName == "" {
			dialog.ShowInformation("Info", "Please select a file first", ui.window)
			return
		}

		name := ui.selectedName
		fullPath := filepath.Join(ui.currentDir, name)

		// 检查是否是文件
		info, err := os.Stat(fullPath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to access %s: %w", name, err), ui.window)
			return
		}

		if info.IsDir() {
			dialog.ShowInformation("Info", "Please select a file, not a directory", ui.window)
			return
		}

		// 计算相对路径
		relativePath, err := filepath.Rel(ui.fileManager.GetWorkingDir(), fullPath)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to get relative path: %w", err), ui.window)
			return
		}

		// 确认删除
		dialog.ShowConfirm("Confirm Delete",
			fmt.Sprintf("Are you sure you want to delete the local file: %s?", relativePath),
			func(confirmed bool) {
				if confirmed {
					if err := ui.fileManager.DeleteLocalFile(relativePath); err != nil {
						dialog.ShowError(err, ui.window)
					} else {
						ui.refreshList()
						dialog.ShowInformation("Success", "File deleted successfully", ui.window)
					}
				}
			}, ui.window)
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

// showRemoteFileDialog shows a dialog to select and download remote files
func (ui *AppUI) showRemoteFileDialog() {
	// 获取远程文件列表
	remoteFiles, err := ui.fileManager.ListRemoteFiles()
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to list remote files: %w", err), ui.window)
		return
	}

	if len(remoteFiles) == 0 {
		dialog.ShowInformation("Info", "No remote files found", ui.window)
		return
	}

	// 创建新窗口显示远程文件
	remoteWindow := ui.app.NewWindow("Remote Files")
	remoteWindow.Resize(fyne.NewSize(700, 500))
	remoteWindow.CenterOnScreen()

	// 创建多选文件列表
	selectedFiles := make(map[int]bool)
	var checkBoxes []*widget.Check

	// 创建滚动容器来容纳复选框列表
	content := container.NewVBox()

	for i, fileName := range remoteFiles {
		index := i // 捕获循环变量
		check := widget.NewCheck(fileName, func(checked bool) {
			selectedFiles[index] = checked
		})
		checkBoxes = append(checkBoxes, check)
		content.Add(check)
	}

	scroll := container.NewScroll(content)
	scroll.SetMinSize(fyne.NewSize(650, 300))

	// 创建全选/全不选按钮
	selectAllBtn := widget.NewButton("Select All", func() {
		for i, check := range checkBoxes {
			check.SetChecked(true)
			selectedFiles[i] = true
		}
	})

	deselectAllBtn := widget.NewButton("Deselect All", func() {
		for i, check := range checkBoxes {
			check.SetChecked(false)
			selectedFiles[i] = false
		}
	})

	// 创建下载按钮
	downloadBtn := widget.NewButton("Download Selected", func() {
		// 收集选中的文件
		var filesToDownload []string
		for i, selected := range selectedFiles {
			if selected && i < len(remoteFiles) {
				filesToDownload = append(filesToDownload, remoteFiles[i])
			}
		}

		if len(filesToDownload) == 0 {
			dialog.ShowInformation("Info", "Please select at least one file", remoteWindow)
			return
		}

		remoteWindow.Close()
		ui.runOperation("Download Multiple Files", func(ctx context.Context) error {
			for _, fileName := range filesToDownload {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				if err := ui.fileManager.DownloadSpecificFile(ctx, fileName); err != nil {
					ui.logger.Error("Failed to download file", slog.String("file", fileName), slog.String("error", err.Error()))
					// 继续下载其他文件，不中断整个过程
				}
			}
			ui.refreshList()
			return nil
		})
	})

	cancelBtn := widget.NewButton("Cancel", func() {
		remoteWindow.Close()
	})

	// 布局
	topButtons := container.NewHBox(selectAllBtn, deselectAllBtn)
	bottomButtons := container.NewHBox(downloadBtn, cancelBtn)

	finalContent := container.NewBorder(
		container.NewVBox(
			widget.NewLabel("Select remote files to download:"),
			topButtons,
		),
		bottomButtons,
		nil,
		nil,
		scroll,
	)

	remoteWindow.SetContent(finalContent)
	remoteWindow.Show()
}

// Run starts the application
func (ui *AppUI) Run() {
	ui.window.ShowAndRun()
}
