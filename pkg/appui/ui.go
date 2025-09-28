package appui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
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
	ui.rightClickableList = NewRightClickableList()
	ui.rightClickableList.OnItemTapped = func(i int) {
		ui.selectedIndex = i
		ui.selectedName = ui.items[i]
		ui.logger.Debug("left click", slog.String("item", ui.selectedName))
	}
	ui.rightClickableList.OnItemRightClick = func(i int, pos fyne.Position) {
		ui.selectedIndex = i
		ui.selectedName = ui.items[i]
		ui.logger.Debug("right click", slog.String("item", ui.selectedName))
		ui.showContextMenu(pos)
	}
	ui.rightClickableList.SetItems(ui.items)
	ui.rightClickableList.Build()

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
	mainContent.SetOffset(ListPaneRatio)

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

	fyne.Do(func() {
		if ui.rightClickableList != nil {
			ui.rightClickableList.SetItems(ui.items)
			ui.rightClickableList.Refresh()
			// 清除选择状态
			ui.rightClickableList.UnselectAll()
		}
		ui.selectedIndex = -1
		ui.selectedName = ""
	})
}

func (ui *AppUI) showContextMenu(pos fyne.Position) {
	if ui.selectedIndex < 0 || ui.selectedIndex >= len(ui.items) {
		return
	}
	menu := fyne.NewMenu("", fyne.NewMenuItem("open in files", ui.openSelectedInFileManager))
	popup := widget.NewPopUpMenu(menu, ui.window.Canvas())
	popup.ShowAtPosition(pos)
}

// goUpDirectory navigates to the parent directory
func (ui *AppUI) goUpDirectory() {
	// 清理当前路径
	cleanCurrentDir := filepath.Clean(ui.currentDir)
	cleanWorkingDir := filepath.Clean(ui.fileManager.GetWorkingDir())

	// 不能超出workingDir的范围
	if cleanCurrentDir == cleanWorkingDir {
		ShowDialog(InfoDialog, ui.window, "Info", "Already at working directory root")
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
		ShowDialog(InfoDialog, ui.window, "Info", "Please select a directory first")
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
		ShowDialogError(fmt.Errorf("failed to access %s: %w", dirName, err), ui.window)
		return
	}

	if !info.IsDir() {
		ShowDialog(InfoDialog, ui.window, "Info", "Selected item is not a directory")
		return
	}

	// 清理路径并确保不会超出workingDir的范围
	cleanFullPath := filepath.Clean(fullPath)
	cleanWorkingDir := filepath.Clean(ui.fileManager.GetWorkingDir())

	// 使用相对路径检查是否在workingDir范围内
	relPath, err := filepath.Rel(cleanWorkingDir, cleanFullPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		ShowDialog(InfoDialog, ui.window, "Info", "Cannot navigate outside working directory")
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
			ShowDialog(InfoDialog, ui.window, "Info", "Please select a file or directory first")
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
			ShowDialog(InfoDialog, ui.window, "Info", "Please select a file first")
			return
		}

		name := ui.selectedName
		fullPath := filepath.Join(ui.currentDir, name)

		// 检查是否是文件
		info, err := os.Stat(fullPath)
		if err != nil {
			ShowDialogError(fmt.Errorf("failed to access %s: %w", name, err), ui.window)
			return
		}

		if info.IsDir() {
			ShowDialog(InfoDialog, ui.window, "Info", "Please select a file, not a directory")
			return
		}

		// 计算相对路径
		relativePath, err := filepath.Rel(ui.fileManager.GetWorkingDir(), fullPath)
		if err != nil {
			ShowDialogError(fmt.Errorf("failed to get relative path: %w", err), ui.window)
			return
		}

		// 确认删除
		dialog.ShowConfirm("Confirm Delete",
			fmt.Sprintf("Are you sure you want to delete the local file: %s?", relativePath),
			func(confirmed bool) {
				if confirmed {
					if err := ui.fileManager.DeleteLocalFile(relativePath); err != nil {
						ShowDialogError(err, ui.window)
					} else {
						ui.refreshList()
						ShowDialog(InfoDialog, ui.window, "Success", "File deleted successfully")
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
			if e := recover(); e != nil {
				ui.logger.Error("Operation panicked", slog.String("operation", operationName))
			}
			// 强制刷新日志文件
			if f, ok := ui.logger.Handler().(interface{ Sync() error }); ok {
				f.Sync()
			}
			fyne.Do(func() {
				ui.operationMutex.Unlock()
			})
		}()

		ui.logger.Info("Starting operation", slog.String("operation", operationName))

		if err := operation(ctx); err != nil {
			if err == context.Canceled {
				ui.logger.Info("Operation cancelled", slog.String("operation", operationName))
			} else {
				ui.logger.Error("Operation failed",
					slog.String("operation", operationName),
					slog.String("error", err.Error()))
				ShowDialogError(err, ui.window)
			}
			return
		}

		ui.logger.Info("Operation completed successfully", slog.String("operation", operationName))
	}()
}

// showRemoteFileDialog shows a dialog to select and download remote files
func (ui *AppUI) showRemoteFileDialog() {
	// 创建新窗口显示远程文件
	remoteWindow := ui.app.NewWindow("Remote Files")
	remoteWindow.Resize(fyne.NewSize(RemoteWindowWidth, RemoteWindowHeight))
	remoteWindow.CenterOnScreen()

	// 创建加载指示器
	loadingLabel := widget.NewLabel("Loading remote files...")
	progressBar := widget.NewProgressBarInfinite()
	progressBar.Start()

	loadingContent := container.NewVBox(
		loadingLabel,
		progressBar,
		widget.NewButton("Cancel", func() {
			remoteWindow.Close()
		}),
	)

	remoteWindow.SetContent(container.NewCenter(loadingContent))
	remoteWindow.Show()

	// 异步获取远程文件列表
	go func() {
		rel, err := filepath.Rel(ui.fileManager.GetWorkingDir(), ui.currentDir)
		if err != nil {
			ui.logger.Warn("Rel path failed", slog.String("workDir", ui.fileManager.GetWorkingDir()), slog.String("currentDir", ui.currentDir))
			rel = ""
		}
		if rel == "." {
			rel = ""
		}

		remoteFiles, err := ui.fileManager.ListRemoteFiles(rel)
		if err != nil {
			fyne.Do(func() {
				remoteWindow.Close()
				ShowDialogError(fmt.Errorf("failed to list remote files: %w", err), ui.window)
			})
			return
		}

		if len(remoteFiles) == 0 {
			fyne.Do(func() {
				remoteWindow.Close()
				ShowDialog(InfoDialog, ui.window, "Info", "No remote files found")
			})
			return
		}

		// 在UI线程中创建文件选择界面
		fyne.Do(func() {
			ui.createRemoteFileSelectionUI(remoteWindow, remoteFiles)
		})
	}()
}

// createRemoteFileSelectionUI creates the file selection interface with performance optimizations
func (ui *AppUI) createRemoteFileSelectionUI(remoteWindow fyne.Window, remoteFiles []string) {
	selectedFiles := make(map[int]bool)

	// 创建搜索框
	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search files...")

	// 过滤后的文件列表
	var filteredFiles []string
	var filteredIndices []int

	// 更新过滤列表的函数
	updateFilteredFiles := func(searchText string) {
		filteredFiles = filteredFiles[:0]
		filteredIndices = filteredIndices[:0]

		searchLower := strings.ToLower(searchText)
		for i, fileName := range remoteFiles {
			if searchText == "" || strings.Contains(strings.ToLower(fileName), searchLower) {
				filteredFiles = append(filteredFiles, fileName)
				filteredIndices = append(filteredIndices, i)
			}
		}
	}

	// 初始化显示所有文件
	updateFilteredFiles("")

	// 创建虚拟化列表容器
	listContainer := container.NewVBox()
	scroll := container.NewScroll(listContainer)
	scroll.SetMinSize(fyne.NewSize(RemoteScrollMinWidth, RemoteScrollMinHeight))

	// 批处理参数
	const batchSize = 50
	var currentBatch int
	var checkBoxes []*widget.Check

	// 状态标签
	statusLabel := widget.NewLabel(fmt.Sprintf("Showing %d files", len(filteredFiles)))

	// 加载更多按钮 - 需要在renderBatch函数之前声明
	var loadMoreBtn *widget.Button

	// 渲染批次的函数
	renderBatch := func() {
		start := currentBatch * batchSize
		end := start + batchSize
		if end > len(filteredFiles) {
			end = len(filteredFiles)
		}

		if start >= len(filteredFiles) {
			return
		}

		for i := start; i < end; i++ {
			fileName := filteredFiles[i]
			originalIndex := filteredIndices[i]

			check := widget.NewCheck(fileName, func(checked bool) {
				selectedFiles[originalIndex] = checked
			})

			// 恢复之前的选择状态
			if selected, exists := selectedFiles[originalIndex]; exists {
				check.SetChecked(selected)
			}

			checkBoxes = append(checkBoxes, check)
			listContainer.Add(check)
		}

		currentBatch++
		listContainer.Refresh()

		// 检查是否需要隐藏加载更多按钮
		if loadMoreBtn != nil && currentBatch*batchSize >= len(filteredFiles) {
			loadMoreBtn.Hide()
		}
	}

	// 重新渲染列表的函数
	rerenderList := func() {
		// 清空现有内容
		listContainer.RemoveAll()
		checkBoxes = checkBoxes[:0]
		currentBatch = 0

		// 更新状态标签
		statusLabel.SetText(fmt.Sprintf("Showing %d files", len(filteredFiles)))

		// 渲染第一批
		if len(filteredFiles) > 0 {
			renderBatch()
		}

		// 显示或隐藏加载更多按钮
		if loadMoreBtn != nil {
			if len(filteredFiles) <= batchSize {
				loadMoreBtn.Hide()
			} else {
				loadMoreBtn.Show()
			}
		}
	}

	// 创建加载更多按钮
	loadMoreBtn = widget.NewButton("Load More", func() {
		renderBatch()
	})

	// 搜索框事件
	searchEntry.OnChanged = func(text string) {
		updateFilteredFiles(text)
		rerenderList()
	}

	// 初始渲染
	rerenderList()

	// 创建全选/全不选按钮
	selectAllBtn := widget.NewButton("Select All", func() {
		for _, index := range filteredIndices {
			selectedFiles[index] = true
		}
		// 更新已渲染的复选框
		for _, check := range checkBoxes {
			check.SetChecked(true)
		}
	})

	deselectAllBtn := widget.NewButton("Deselect All", func() {
		for _, index := range filteredIndices {
			selectedFiles[index] = false
		}
		// 更新已渲染的复选框
		for _, check := range checkBoxes {
			check.SetChecked(false)
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
			ShowDialog(InfoDialog, remoteWindow, "Info", "Please select at least one file")
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
	topSection := container.NewVBox(
		widget.NewLabel("Select remote files to download:"),
		searchEntry,
		statusLabel,
		container.NewHBox(selectAllBtn, deselectAllBtn),
	)

	bottomSection := container.NewVBox(
		loadMoreBtn,
		container.NewHBox(downloadBtn, cancelBtn),
	)

	finalContent := container.NewBorder(
		topSection,
		bottomSection,
		nil,
		nil,
		scroll,
	)

	remoteWindow.SetContent(finalContent)
}

// GetLogWidget returns the log widget for setting up log handler
func (ui *AppUI) GetLogWidget() *widget.TextGrid {
	return ui.logWidget
}

// openSelectedInFileManager opens the file manager for the currently selected item
func (ui *AppUI) openSelectedInFileManager() {
	if ui.selectedIndex < 0 || ui.selectedIndex >= len(ui.items) {
		ShowDialog(InfoDialog, ui.window, "Info", "Please select a file or directory first")
		return
	}
	fullPath := filepath.Join(ui.currentDir, ui.selectedName)
	if err := ui.openInFileManager(fullPath); err != nil {
		ui.logger.Error("Failed to open file manager", slog.String("error", err.Error()))
		ShowDialogError(fmt.Errorf("failed to open file manager: %w", err), ui.window)
	}
	ui.selectedIndex = -1
	ui.selectedName = ""
	ui.rightClickableList.UnselectAll()
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
	defer func() {
		if r := recover(); r != nil {
			ui.logger.Error("UI Run panic",
				slog.String("error", fmt.Sprintf("%v", r)),
				slog.String("stack", string(debug.Stack())))
		}
	}()
	ui.window.ShowAndRun()
}
