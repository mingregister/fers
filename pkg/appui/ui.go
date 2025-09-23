package appui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/mingregister/fers/pkg/dir"
)

// UI Constants
const (
	DefaultWindowWidth    = 1000
	DefaultWindowHeight   = 600
	ListPaneRatio         = 0.8 // 80% for file list, 20% for logs
	LogPaneMinWidth       = 400
	LogPaneMinHeight      = 200
	RemoteWindowWidth     = 700
	RemoteWindowHeight    = 500
	RemoteScrollMinWidth  = 650
	RemoteScrollMinHeight = 300
)

// ItemContainer wraps each list item to handle right-click events
type ItemContainer struct {
	widget.BaseWidget
	label *widget.Label
	rcl   *RightClickableList
	index int
}

// CreateRenderer implements fyne.Widget interface for ItemContainer
func (ic *ItemContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ic.label)
}

// selectItem updates the selection to the specified item
func (ic *ItemContainer) selectItem() {
	// Boundary check
	if ic.index < 0 || ic.index >= len(ic.rcl.ui.items) {
		ic.rcl.ui.logger.Debug("Invalid item index", slog.Int("index", ic.index), slog.Int("total_items", len(ic.rcl.ui.items)))
		return
	}

	ic.rcl.ui.selectedIndex = ic.index
	ic.rcl.ui.selectedName = ic.rcl.ui.items[ic.index]
	ic.rcl.list.Select(ic.index)
}

// Tapped handles left-click on individual items
func (ic *ItemContainer) Tapped(pe *fyne.PointEvent) {
	ic.rcl.ui.logger.Debug("ItemContainer Tapped called!", slog.Int("index", ic.index))

	ic.selectItem()
	ic.rcl.ui.logger.Debug("Selected item via left-click", slog.String("item", ic.rcl.ui.selectedName))
}

// TappedSecondary handles right-click on individual items
func (ic *ItemContainer) TappedSecondary(pe *fyne.PointEvent) {
	ic.rcl.ui.logger.Debug("ItemContainer TappedSecondary called!", slog.Int("index", ic.index))

	ic.selectItem()
	ic.rcl.ui.logger.Debug("Selected item via right-click", slog.String("item", ic.rcl.ui.selectedName))

	// Show context menu
	ic.rcl.showContextMenu(pe.AbsolutePosition)
}

// RightClickableList is a custom widget that wraps widget.List with right-click support
type RightClickableList struct {
	widget.BaseWidget
	list *widget.List
	ui   *AppUI
}

// Compile-time interface checks
var _ fyne.Widget = (*RightClickableList)(nil)
var _ fyne.Widget = (*ItemContainer)(nil)
var _ fyne.Tappable = (*ItemContainer)(nil)
var _ fyne.SecondaryTappable = (*ItemContainer)(nil)

// NewRightClickableList creates a new RightClickableList with right-click support
func NewRightClickableList(ui *AppUI) *RightClickableList {
	rcl := &RightClickableList{ui: ui}

	// Create internal list
	rcl.list = widget.NewList(
		func() int { return len(ui.items) },
		func() fyne.CanvasObject {
			// Create a custom widget that can handle right-click for each item
			label := widget.NewLabel("template")
			container := &ItemContainer{
				label: label,
				rcl:   rcl,
			}
			// Extend BaseWidget - crucial for Fyne to recognize this as a widget
			container.ExtendBaseWidget(container)
			return container
		},
		func(i int, o fyne.CanvasObject) {
			itemContainer := o.(*ItemContainer)
			itemContainer.label.SetText(ui.items[i])
			itemContainer.index = i
		},
	)

	// Set up selection handler
	rcl.list.OnSelected = func(i int) {
		ui.selectedIndex = i
		ui.selectedName = ui.items[i]
	}

	// Extend BaseWidget - this is crucial for Fyne to recognize this as a widget
	rcl.ExtendBaseWidget(rcl)

	// Debug: Log that we created the widget
	ui.logger.Info("Created RightClickableList widget")

	return rcl
}

// CreateRenderer implements fyne.Widget interface - this is called by Fyne framework
func (rcl *RightClickableList) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(rcl.list)
}

// showContextMenu displays the right-click context menu
func (rcl *RightClickableList) showContextMenu(pos fyne.Position) {
	if rcl.ui.selectedIndex < 0 || rcl.ui.selectedIndex >= len(rcl.ui.items) {
		return
	}

	menu := fyne.NewMenu("", fyne.NewMenuItem("open in files", rcl.ui.openSelectedInFileManager))
	popup := widget.NewPopUpMenu(menu, rcl.ui.window.Canvas())
	popup.ShowAtPosition(pos)
	popup.Show()
}

// GetList returns the underlying list widget for direct access if needed
func (rcl *RightClickableList) GetList() *widget.List {
	return rcl.list
}

// Refresh refreshes the list display
func (rcl *RightClickableList) Refresh() {
	rcl.list.Refresh()
}

// UnselectAll clears all selections
func (rcl *RightClickableList) UnselectAll() {
	rcl.list.UnselectAll()
}

// AppUI manages the user interface
type AppUI struct {
	app         fyne.App
	window      fyne.Window
	fileManager *dir.FileManager
	logger      *slog.Logger

	// UI components
	rightClickableList *RightClickableList
	items              []string
	selectedIndex      int
	selectedName       string
	logWidget          *widget.TextGrid

	// Directory navigation
	currentDir string // 当前显示的目录
	dirLabel   *widget.Label

	// Operation management
	operationMutex sync.Mutex
	cancelFunc     context.CancelFunc
}

// validateSelection checks if a valid item is selected
func (ui *AppUI) validateSelection() bool {
	return ui.selectedIndex >= 0 && ui.selectedIndex < len(ui.items) && ui.selectedName != ""
}

// NewAppUI creates a new AppUI instance
func NewAppUI(fileManager *dir.FileManager, logger *slog.Logger) *AppUI {
	app := app.New()
	window := app.NewWindow("File Encrypt & Remote Storage")
	window.Resize(fyne.NewSize(DefaultWindowWidth, DefaultWindowHeight))
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

// NewAppUIWithLogWidget creates a new AppUI instance with a pre-created log widget
func NewAppUIWithLogWidget(fileManager *dir.FileManager, logger *slog.Logger, logWidget *widget.TextGrid) *AppUI {
	app := app.New()
	window := app.NewWindow("File Encrypt & Remote Storage")
	window.Resize(fyne.NewSize(DefaultWindowWidth, DefaultWindowHeight))
	window.CenterOnScreen()

	ui := &AppUI{
		app:           app,
		window:        window,
		fileManager:   fileManager,
		logger:        logger,
		selectedIndex: -1,
		currentDir:    fileManager.GetWorkingDir(), // 初始化为workingDir
		logWidget:     logWidget,
	}

	ui.setupUI()
	return ui
}

// setupUI initializes the user interface
func (ui *AppUI) setupUI() {
	// Directory labels
	workingDirLabel := widget.NewLabel("Working dir: " + ui.fileManager.GetWorkingDir())
	ui.dirLabel = widget.NewLabel("Current dir: " + ui.currentDir)

	// File list with right-click support
	ui.refreshItems()
	ui.rightClickableList = NewRightClickableList(ui)

	// Log widget - create only if not already provided
	if ui.logWidget == nil {
		ui.logWidget = widget.NewTextGrid()
		ui.logWidget.SetText("Application Logs\n\nLogs will appear here...\n")
	}
	logScroll := container.NewScroll(ui.logWidget)
	logScroll.SetMinSize(fyne.NewSize(LogPaneMinWidth, LogPaneMinHeight))

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

	// Layout - directly use the custom widget
	dirLabels := container.NewVBox(workingDirLabel, ui.dirLabel)
	ListPane := container.NewBorder(dirLabels, nil, nil, nil, ui.rightClickableList)

	// Create main content with file list on left and log on right
	mainContent := container.NewVSplit(ListPane, logScroll)
	mainContent.SetOffset(0.8) // 80% for file list, 40% for logs

	content := container.NewBorder(nil, nil, buttons, nil, mainContent)
	ui.window.SetContent(content)
}

// refreshItems updates the items list
func (ui *AppUI) refreshItems() {
	ui.items = dir.List(ui.currentDir)
}

// refreshList refreshes the UI list
func (ui *AppUI) refreshList() {
	ui.refreshItems()
	if ui.rightClickableList != nil {
		ui.rightClickableList.Refresh()
		// 清除选择状态
		ui.rightClickableList.UnselectAll()
	}
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
	ui.dirLabel.SetText(fmt.Sprintf("Current dir: %s", ui.currentDir))
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
		if !ui.validateSelection() {
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
		if !ui.validateSelection() {
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
	remoteWindow.Resize(fyne.NewSize(RemoteWindowWidth, RemoteWindowHeight))
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
	scroll.SetMinSize(fyne.NewSize(RemoteScrollMinWidth, RemoteScrollMinHeight))

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

// GetLogWidget returns the log widget for setting up log handler
func (ui *AppUI) GetLogWidget() *widget.TextGrid {
	return ui.logWidget
}

// openSelectedInFileManager opens the file manager for the currently selected item
func (ui *AppUI) openSelectedInFileManager() {
	if !ui.validateSelection() {
		dialog.ShowInformation("Info", "Please select a file or directory first", ui.window)
		return
	}

	// Get the full path of the selected item
	fullPath := filepath.Join(ui.currentDir, ui.selectedName)

	// Open the file manager
	if err := ui.openInFileManager(fullPath); err != nil {
		ui.logger.Error("Failed to open file manager", slog.String("error", err.Error()))
		dialog.ShowError(fmt.Errorf("failed to open file manager: %w", err), ui.window)
	} else {
		// Reset selection after successful operation
		ui.selectedIndex = -1
		ui.selectedName = ""
		if ui.rightClickableList != nil {
			ui.rightClickableList.UnselectAll()
		}
		ui.logger.Info("Selection cleared after opening file manager")
	}
}

// openInFileManager opens the system file manager at the specified path
func (ui *AppUI) openInFileManager(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Use explorer with /select to highlight the file/folder
		windowsPath := filepath.Clean(path)
		cmd = exec.Command("explorer", "/select,"+windowsPath)
	case "darwin":
		// Use open with -R to reveal in Finder
		cmd = exec.Command("open", "-R", path)
	case "linux":
		// Check if the path is a directory or file
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to stat path %s: %w", path, err)
		}

		if info.IsDir() {
			// Open the directory directly
			cmd = exec.Command("xdg-open", path)
		} else {
			// Open the parent directory
			parentDir := filepath.Dir(path)
			cmd = exec.Command("xdg-open", parentDir)
		}
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	ui.logger.Info("Opening file manager", slog.String("path", path), slog.String("os", runtime.GOOS))

	// Windows 用 Start()，其他系统用 Run()
	if runtime.GOOS == "windows" {
		return cmd.Start()
	}
	return cmd.Run()
}

// Run starts the application
func (ui *AppUI) Run() {
	ui.window.ShowAndRun()
}
