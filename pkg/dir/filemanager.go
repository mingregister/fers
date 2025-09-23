package dir

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/mingregister/fers/pkg/config"
	"github.com/mingregister/fers/pkg/crypto"
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
	cipher     crypto.Cipher
	logger     *slog.Logger
}

// NewFileManager creates a new FileManager instance
func NewFileManager(cfg *config.Config, storage storage.Client, logger *slog.Logger, cipher crypto.Cipher) *FileManager {
	return &FileManager{
		config:     cfg,
		storage:    storage,
		workingDir: cfg.TargetDir,
		cipher:     cipher,
		logger:     logger,
	}
}

func (fm *FileManager) GetWorkingDir() string {
	return fm.workingDir
}

// EncryptAndUploadFile encrypts and uploads a single file
func (fm *FileManager) EncryptAndUploadFile(filePath, relativePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	encrypted, err := fm.cipher.Encrypt(data)
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

	decrypted, err := fm.cipher.Decrypt(encrypted)
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

	// 构建本地文件的完整路径集合
	localFileSet := make(map[string]bool)
	err = filepath.Walk(fm.workingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relativePath, err := filepath.Rel(fm.workingDir, path)
			if err != nil {
				return err
			}
			// 使用斜杠路径以匹配远程路径格式
			localFileSet[filepath.ToSlash(relativePath)] = true
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to scan local files: %w", err)
	}

	for _, remotePath := range remoteFiles {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 检查远程文件是否在本地存在
		if !localFileSet[remotePath] {
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

// ListRemoteFiles returns a list of all remote files
func (fm *FileManager) ListRemoteFiles(prefix string) ([]string, error) {
	return fm.storage.List(prefix)
}

// DownloadSpecificFile downloads a specific file from remote storage
func (fm *FileManager) DownloadSpecificFile(ctx context.Context, remotePath string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	localPath := filepath.Join(fm.workingDir, remotePath)
	return fm.DownloadAndDecryptFile(remotePath, localPath)
}

// DeleteLocalFile deletes a local file
func (fm *FileManager) DeleteLocalFile(relativePath string) error {
	localPath := filepath.Join(fm.workingDir, relativePath)

	// 确保文件在工作目录范围内
	cleanLocalPath := filepath.Clean(localPath)
	cleanWorkingDir := filepath.Clean(fm.workingDir)

	relPath, err := filepath.Rel(cleanWorkingDir, cleanLocalPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("file is outside working directory")
	}

	if err := os.Remove(localPath); err != nil {
		return fmt.Errorf("failed to delete file %s: %w", relativePath, err)
	}

	fm.logger.Info("File deleted successfully", slog.String("path", relativePath))
	return nil
}
