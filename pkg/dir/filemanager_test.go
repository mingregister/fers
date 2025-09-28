package dir

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mingregister/fers/pkg/config"
	"github.com/mingregister/fers/pkg/crypto"
)

// Mock implementations for testing
type mockStorage struct {
	files map[string][]byte
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		files: make(map[string][]byte),
	}
}

func (m *mockStorage) List(prefix string) ([]string, error) {
	var result []string
	for key := range m.files {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			result = append(result, key)
		}
	}
	return result, nil
}

func (m *mockStorage) Upload(key string, data []byte) error {
	m.files[key] = data
	return nil
}

func (m *mockStorage) Download(key string) ([]byte, error) {
	data, exists := m.files[key]
	if !exists {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (m *mockStorage) Delete(key string) error {
	delete(m.files, key)
	return nil
}

func createTestFileManager(t *testing.T) (*FileManager, string, *mockStorage) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		TargetDir: tempDir,
		CryptoKey: "test-key-123",
	}

	mockStore := newMockStorage()
	cipher := crypto.NewAESGCM("test-password")
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	fm := NewFileManager(cfg, mockStore, logger, cipher)

	return fm, tempDir, mockStore
}

func TestNewFileManager(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{TargetDir: tempDir}
	mockStore := newMockStorage()
	cipher := crypto.NewAESGCM("test")
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	fm := NewFileManager(cfg, mockStore, logger, cipher)

	if fm == nil {
		t.Fatal("NewFileManager returned nil")
	}

	if fm.GetWorkingDir() != tempDir {
		t.Errorf("Expected working dir %s, got %s", tempDir, fm.GetWorkingDir())
	}
}

func TestFileManager_EncryptAndUploadFile(t *testing.T) {
	fm, tempDir, mockStore := createTestFileManager(t)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("hello world")
	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Upload the file
	err = fm.EncryptAndUploadFile(testFile, "test.txt")
	if err != nil {
		t.Fatalf("EncryptAndUploadFile failed: %v", err)
	}

	// Verify file was uploaded
	uploadedData, exists := mockStore.files["test.txt"]
	if !exists {
		t.Fatal("File was not uploaded to storage")
	}

	// Verify data is encrypted (should be different from original)
	if string(uploadedData) == string(testContent) {
		t.Error("Uploaded data should be encrypted")
	}

	// Verify we can decrypt it back
	cipher := crypto.NewAESGCM("test-password")
	decrypted, err := cipher.Decrypt(uploadedData)
	if err != nil {
		t.Fatalf("Failed to decrypt uploaded data: %v", err)
	}

	if string(decrypted) != string(testContent) {
		t.Errorf("Decrypted content mismatch. Expected: %s, Got: %s", testContent, decrypted)
	}
}

func TestFileManager_EncryptAndUploadDirectory(t *testing.T) {
	fm, tempDir, mockStore := createTestFileManager(t)

	// Create test directory structure
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create test files
	files := map[string]string{
		"file1.txt":        "content1",
		"file2.txt":        "content2",
		"subdir/file3.txt": "content3",
	}

	for relPath, content := range files {
		fullPath := filepath.Join(tempDir, relPath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Upload directory
	ctx := context.Background()
	err = fm.EncryptAndUploadDirectory(ctx, tempDir)
	if err != nil {
		t.Fatalf("EncryptAndUploadDirectory failed: %v", err)
	}

	// Verify all files were uploaded
	if len(mockStore.files) != len(files) {
		t.Errorf("Expected %d files uploaded, got %d", len(files), len(mockStore.files))
	}

	// Verify each file
	cipher := crypto.NewAESGCM("test-password")
	for relPath, expectedContent := range files {
		uploadedData, exists := mockStore.files[filepath.ToSlash(relPath)]
		if !exists {
			t.Errorf("File %s was not uploaded", relPath)
			continue
		}

		decrypted, err := cipher.Decrypt(uploadedData)
		if err != nil {
			t.Errorf("Failed to decrypt %s: %v", relPath, err)
			continue
		}

		if string(decrypted) != expectedContent {
			t.Errorf("Content mismatch for %s. Expected: %s, Got: %s", relPath, expectedContent, decrypted)
		}
	}
}

func TestFileManager_DownloadAndDecryptFile(t *testing.T) {
	fm, tempDir, mockStore := createTestFileManager(t)

	// Prepare encrypted data in mock storage
	cipher := crypto.NewAESGCM("test-password")
	originalContent := []byte("test download content")
	encrypted, err := cipher.Encrypt(originalContent)
	if err != nil {
		t.Fatalf("Failed to encrypt test data: %v", err)
	}

	mockStore.files["download/test.txt"] = encrypted

	// Download and decrypt
	localPath := filepath.Join(tempDir, "downloaded.txt")
	err = fm.DownloadAndDecryptFile("download/test.txt", localPath)
	if err != nil {
		t.Fatalf("DownloadAndDecryptFile failed: %v", err)
	}

	// Verify file was created and content is correct
	content, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != string(originalContent) {
		t.Errorf("Downloaded content mismatch. Expected: %s, Got: %s", originalContent, content)
	}
}

func TestFileManager_SyncDownload(t *testing.T) {
	fm, tempDir, mockStore := createTestFileManager(t)

	// Prepare some files in remote storage
	cipher := crypto.NewAESGCM("test-password")
	remoteFiles := map[string]string{
		"remote1.txt":        "remote content 1",
		"folder/remote2.txt": "remote content 2",
		"existing.txt":       "existing content",
	}

	for remotePath, content := range remoteFiles {
		encrypted, err := cipher.Encrypt([]byte(content))
		if err != nil {
			t.Fatalf("Failed to encrypt %s: %v", remotePath, err)
		}
		mockStore.files[remotePath] = encrypted
	}

	// Create one file locally that already exists remotely
	existingFile := filepath.Join(tempDir, "existing.txt")
	err := os.WriteFile(existingFile, []byte("local existing content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Sync download
	ctx := context.Background()
	err = fm.SyncDownload(ctx)
	if err != nil {
		t.Fatalf("SyncDownload failed: %v", err)
	}

	// Verify missing files were downloaded
	expectedDownloads := []string{"remote1.txt", "folder/remote2.txt"}
	for _, relPath := range expectedDownloads {
		localPath := filepath.Join(tempDir, relPath)
		if _, err := os.Stat(localPath); os.IsNotExist(err) {
			t.Errorf("File %s was not downloaded", relPath)
			continue
		}

		content, err := os.ReadFile(localPath)
		if err != nil {
			t.Errorf("Failed to read downloaded file %s: %v", relPath, err)
			continue
		}

		expectedContent := remoteFiles[relPath]
		if string(content) != expectedContent {
			t.Errorf("Content mismatch for %s. Expected: %s, Got: %s", relPath, expectedContent, content)
		}
	}

	// Verify existing file was not overwritten
	existingContent, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read existing file: %v", err)
	}
	if string(existingContent) != "local existing content" {
		t.Error("Existing file was overwritten during sync download")
	}
}

func TestFileManager_SyncUpload(t *testing.T) {
	fm, tempDir, mockStore := createTestFileManager(t)

	// Create local files
	localFiles := map[string]string{
		"local1.txt":        "local content 1",
		"folder/local2.txt": "local content 2",
		"existing.txt":      "local existing content",
	}

	for relPath, content := range localFiles {
		fullPath := filepath.Join(tempDir, relPath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// Add one file that already exists remotely
	cipher := crypto.NewAESGCM("test-password")
	existingEncrypted, err := cipher.Encrypt([]byte("remote existing content"))
	if err != nil {
		t.Fatalf("Failed to encrypt existing remote file: %v", err)
	}
	mockStore.files["existing.txt"] = existingEncrypted

	// Sync upload
	ctx := context.Background()
	err = fm.SyncUpload(ctx)
	if err != nil {
		t.Fatalf("SyncUpload failed: %v", err)
	}

	// Verify missing files were uploaded
	expectedUploads := []string{"local1.txt", "folder/local2.txt"}
	for _, relPath := range expectedUploads {
		uploadedData, exists := mockStore.files[filepath.ToSlash(relPath)]
		if !exists {
			t.Errorf("File %s was not uploaded", relPath)
			continue
		}

		decrypted, err := cipher.Decrypt(uploadedData)
		if err != nil {
			t.Errorf("Failed to decrypt uploaded %s: %v", relPath, err)
			continue
		}

		expectedContent := localFiles[relPath]
		if string(decrypted) != expectedContent {
			t.Errorf("Uploaded content mismatch for %s. Expected: %s, Got: %s", relPath, expectedContent, decrypted)
		}
	}

	// Verify existing remote file was not overwritten
	existingData := mockStore.files["existing.txt"]
	decrypted, err := cipher.Decrypt(existingData)
	if err != nil {
		t.Fatalf("Failed to decrypt existing remote file: %v", err)
	}
	if string(decrypted) != "remote existing content" {
		t.Error("Existing remote file was overwritten during sync upload")
	}
}

func TestFileManager_ListRemoteFiles(t *testing.T) {
	fm, _, mockStore := createTestFileManager(t)

	// Add files to mock storage
	mockStore.files["file1.txt"] = []byte("data1")
	mockStore.files["folder/file2.txt"] = []byte("data2")
	mockStore.files["folder/file3.txt"] = []byte("data3")
	mockStore.files["other/file4.txt"] = []byte("data4")

	// Test listing all files
	files, err := fm.ListRemoteFiles("")
	if err != nil {
		t.Fatalf("ListRemoteFiles failed: %v", err)
	}

	if len(files) != 4 {
		t.Errorf("Expected 4 files, got %d", len(files))
	}

	// Test listing with prefix
	folderFiles, err := fm.ListRemoteFiles("folder/")
	if err != nil {
		t.Fatalf("ListRemoteFiles with prefix failed: %v", err)
	}

	if len(folderFiles) != 2 {
		t.Errorf("Expected 2 files with folder/ prefix, got %d", len(folderFiles))
	}
}

func TestFileManager_DownloadSpecificFile(t *testing.T) {
	fm, tempDir, mockStore := createTestFileManager(t)

	// Prepare encrypted data in mock storage
	cipher := crypto.NewAESGCM("test-password")
	originalContent := []byte("specific file content")
	encrypted, err := cipher.Encrypt(originalContent)
	if err != nil {
		t.Fatalf("Failed to encrypt test data: %v", err)
	}

	mockStore.files["specific/file.txt"] = encrypted

	// Download specific file
	ctx := context.Background()
	err = fm.DownloadSpecificFile(ctx, "specific/file.txt")
	if err != nil {
		t.Fatalf("DownloadSpecificFile failed: %v", err)
	}

	// Verify file was downloaded
	localPath := filepath.Join(tempDir, "specific", "file.txt")
	content, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != string(originalContent) {
		t.Errorf("Downloaded content mismatch. Expected: %s, Got: %s", originalContent, content)
	}
}

func TestFileManager_DeleteLocalFile(t *testing.T) {
	fm, tempDir, _ := createTestFileManager(t)

	// Create a test file
	testFile := filepath.Join(tempDir, "delete_me.txt")
	err := os.WriteFile(testFile, []byte("delete this"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Delete the file
	err = fm.DeleteLocalFile("delete_me.txt")
	if err != nil {
		t.Fatalf("DeleteLocalFile failed: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File was not deleted")
	}
}

func TestFileManager_DeleteLocalFile_SecurityCheck(t *testing.T) {
	fm, _, _ := createTestFileManager(t)

	// Try to delete a file outside working directory
	err := fm.DeleteLocalFile("../outside.txt")
	if err == nil {
		t.Error("Should not allow deleting files outside working directory")
	}

	// Try to delete with absolute path outside working directory
	err = fm.DeleteLocalFile("/etc/passwd")
	if err == nil {
		t.Error("Should not allow deleting files with absolute paths outside working directory")
	}
}

func TestFileManager_ContextCancellation(t *testing.T) {
	fm, tempDir, _ := createTestFileManager(t)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Try to upload directory with cancelled context
	err = fm.EncryptAndUploadDirectory(ctx, tempDir)
	if err == nil {
		t.Error("Expected error for cancelled context, got nil")
	} else if err != context.Canceled {
		// On some systems, the error might be wrapped or different
		t.Logf("Got error (expected context cancellation): %v", err)
	}

	// // Try sync operations with cancelled context
	// err = fm.SyncDownload(ctx)
	// if err == nil {
	// 	t.Error("Expected error for cancelled context in SyncDownload, got nil")
	// } else if err != context.Canceled {
	// 	t.Logf("Got error for SyncDownload (expected context cancellation): %v", err)
	// }

	err = fm.SyncUpload(ctx)
	if err == nil {
		t.Error("Expected error for cancelled context in SyncUpload, got nil")
	} else if err != context.Canceled {
		t.Logf("Got error for SyncUpload (expected context cancellation): %v", err)
	}

	err = fm.DownloadSpecificFile(ctx, "test.txt")
	if err == nil {
		t.Error("Expected error for cancelled context in DownloadSpecificFile, got nil")
	} else if err != context.Canceled {
		t.Logf("Got error for DownloadSpecificFile (expected context cancellation): %v", err)
	}
}

func TestFileManager_ContextTimeout(t *testing.T) {
	fm, tempDir, _ := createTestFileManager(t)

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait a bit to ensure timeout
	time.Sleep(1 * time.Millisecond)

	// Try operations with timed out context
	err = fm.EncryptAndUploadDirectory(ctx, tempDir)
	if err == nil {
		t.Error("Expected timeout error, got nil")
	} else if err != context.DeadlineExceeded {
		// On some systems, the error might be wrapped or different
		t.Logf("Got error (expected timeout): %v", err)
	}
}

func TestFileManager_ErrorHandling(t *testing.T) {
	fm, _, _ := createTestFileManager(t)

	// Test uploading non-existent file (use Windows-compatible path)
	err := fm.EncryptAndUploadFile("C:\\nonexistent\\file.txt", "test.txt")
	if err == nil {
		t.Error("Should fail when uploading non-existent file")
	}

	// Test downloading to invalid path (use Windows-compatible path)
	err = fm.DownloadAndDecryptFile("test.txt", "C:\\invalid\\path\\file.txt")
	if err == nil {
		t.Error("Should fail when downloading to invalid path")
	}

	// Test deleting non-existent file
	err = fm.DeleteLocalFile("nonexistent.txt")
	if err == nil {
		t.Error("Should fail when deleting non-existent file")
	}
}
