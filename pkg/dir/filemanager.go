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
func NewFileManager(cfg *config.Config, storage storage.Client, logger *slog.Logger) *FileManager {
	return &FileManager{
		config:     cfg,
		storage:    storage,
		workingDir: cfg.TargetDir,
		cipher:     crypto.NewAESGCM(cfg.CryptoKey),
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

	localFiles := List(fm.workingDir)
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
