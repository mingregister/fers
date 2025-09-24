package storage

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNewOSSMock(t *testing.T) {
	tempDir := t.TempDir()

	client := NewOSSMock(tempDir)
	if client == nil {
		t.Fatal("NewOSSMock returned nil")
	}

	// Verify directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Base directory was not created")
	}
}

func TestOSSMock_Upload(t *testing.T) {
	tempDir := t.TempDir()
	client := NewOSSMock(tempDir)

	testCases := []struct {
		name string
		key  string
		data []byte
	}{
		{
			name: "simple file",
			key:  "test.txt",
			data: []byte("hello world"),
		},
		{
			name: "nested path",
			key:  "folder/subfolder/file.txt",
			data: []byte("nested content"),
		},
		{
			name: "binary data",
			key:  "binary.bin",
			data: []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
		},
		{
			name: "empty file",
			key:  "empty.txt",
			data: []byte{},
		},
		{
			name: "unicode filename",
			key:  "测试文件.txt",
			data: []byte("unicode content"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.Upload(tc.key, tc.data)
			if err != nil {
				t.Fatalf("Upload failed: %v", err)
			}

			// Verify file was created
			expectedPath := filepath.Join(tempDir, filepath.FromSlash(tc.key))
			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				t.Error("File was not created")
			}

			// Verify file content
			content, err := os.ReadFile(expectedPath)
			if err != nil {
				t.Fatalf("Failed to read uploaded file: %v", err)
			}

			if !bytes.Equal(content, tc.data) {
				t.Errorf("File content mismatch.\nExpected: %v\nGot: %v", tc.data, content)
			}
		})
	}
}

func TestOSSMock_Download(t *testing.T) {
	tempDir := t.TempDir()
	client := NewOSSMock(tempDir)

	// Upload test data first
	testData := []byte("download test data")
	key := "download/test.txt"

	err := client.Upload(key, testData)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	// Download the data
	downloaded, err := client.Download(key)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	if !bytes.Equal(downloaded, testData) {
		t.Errorf("Downloaded data mismatch.\nExpected: %v\nGot: %v", testData, downloaded)
	}
}

func TestOSSMock_DownloadNonExistent(t *testing.T) {
	tempDir := t.TempDir()
	client := NewOSSMock(tempDir)

	_, err := client.Download("nonexistent.txt")
	if err == nil {
		t.Error("Download should fail for non-existent file")
	}
}

func TestOSSMock_List(t *testing.T) {
	tempDir := t.TempDir()
	client := NewOSSMock(tempDir)

	// Upload test files
	testFiles := map[string][]byte{
		"file1.txt":            []byte("content1"),
		"folder/file2.txt":     []byte("content2"),
		"folder/file3.txt":     []byte("content3"),
		"other/file4.txt":      []byte("content4"),
		"folder/sub/file5.txt": []byte("content5"),
	}

	for key, data := range testFiles {
		err := client.Upload(key, data)
		if err != nil {
			t.Fatalf("Upload failed for %s: %v", key, err)
		}
	}

	testCases := []struct {
		name     string
		prefix   string
		expected []string
	}{
		{
			name:     "list all",
			prefix:   "",
			expected: []string{"file1.txt", "folder/file2.txt", "folder/file3.txt", "folder/sub/file5.txt", "other/file4.txt"},
		},
		{
			name:     "list folder prefix",
			prefix:   "folder/",
			expected: []string{"folder/file2.txt", "folder/file3.txt", "folder/sub/file5.txt"},
		},
		{
			name:     "list other prefix",
			prefix:   "other/",
			expected: []string{"other/file4.txt"},
		},
		{
			name:     "list non-existent prefix",
			prefix:   "nonexistent/",
			expected: []string{},
		},
		{
			name:     "list specific file prefix",
			prefix:   "file1",
			expected: []string{"file1.txt"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			files, err := client.List(tc.prefix)
			if err != nil {
				t.Fatalf("List failed: %v", err)
			}

			// Sort both slices for comparison
			if len(files) != len(tc.expected) {
				t.Errorf("Expected %d files, got %d.\nExpected: %v\nGot: %v", len(tc.expected), len(files), tc.expected, files)
				return
			}

			// Check if all expected files are present
			expectedMap := make(map[string]bool)
			for _, f := range tc.expected {
				expectedMap[f] = true
			}

			for _, f := range files {
				if !expectedMap[f] {
					t.Errorf("Unexpected file in list: %s", f)
				}
				delete(expectedMap, f)
			}

			if len(expectedMap) > 0 {
				var missing []string
				for f := range expectedMap {
					missing = append(missing, f)
				}
				t.Errorf("Missing files in list: %v", missing)
			}
		})
	}
}

func TestOSSMock_ListEmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()
	client := NewOSSMock(tempDir)

	files, err := client.List("")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected empty list, got %v", files)
	}
}

func TestOSSMock_Delete(t *testing.T) {
	tempDir := t.TempDir()
	client := NewOSSMock(tempDir)

	// Note: The current implementation of Delete is a no-op
	// This test verifies that it doesn't return an error
	err := client.Delete("any-key")
	if err != nil {
		t.Errorf("Delete should not return error, got: %v", err)
	}
}

func TestOSSMock_ConcurrentOperations(t *testing.T) {
	tempDir := t.TempDir()
	client := NewOSSMock(tempDir)

	// Test concurrent uploads
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			key := fmt.Sprintf("concurrent/file%d.txt", id)
			data := []byte(fmt.Sprintf("data%d", id))

			err := client.Upload(key, data)
			if err != nil {
				t.Errorf("Concurrent upload failed for %s: %v", key, err)
				return
			}

			// Verify download
			downloaded, err := client.Download(key)
			if err != nil {
				t.Errorf("Concurrent download failed for %s: %v", key, err)
				return
			}

			if !bytes.Equal(downloaded, data) {
				t.Errorf("Concurrent data mismatch for %s", key)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all files were created
	files, err := client.List("concurrent/")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(files) != 10 {
		t.Errorf("Expected 10 files, got %d", len(files))
	}
}

func TestOSSMock_PathSeparators(t *testing.T) {
	tempDir := t.TempDir()
	client := NewOSSMock(tempDir)

	// Test with different path separators
	testCases := []struct {
		key      string
		expected string // expected path in list output (always uses forward slashes)
	}{
		{
			key:      "folder/file.txt",
			expected: "folder/file.txt",
		},
		{
			key:      "deep/nested/path/file.txt",
			expected: "deep/nested/path/file.txt",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			data := []byte("test data")

			err := client.Upload(tc.key, data)
			if err != nil {
				t.Fatalf("Upload failed: %v", err)
			}

			// Verify file appears in list with correct path format
			files, err := client.List("")
			if err != nil {
				t.Fatalf("List failed: %v", err)
			}

			found := false
			for _, f := range files {
				if f == tc.expected {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("File %s not found in list. Got: %v", tc.expected, files)
			}

			// Verify download works
			downloaded, err := client.Download(tc.key)
			if err != nil {
				t.Fatalf("Download failed: %v", err)
			}

			if !bytes.Equal(downloaded, data) {
				t.Error("Downloaded data mismatch")
			}
		})
	}
}

func TestOSSMock_InterfaceCompliance(t *testing.T) {
	tempDir := t.TempDir()
	client := NewOSSMock(tempDir)

	// Test that it implements the Client interface
	var _ Client = client
}
