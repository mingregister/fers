package dir

import (
	"os"
	"path/filepath"
	"testing"
)

func TestList(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test files and directories
	testItems := []struct {
		name   string
		isDir  bool
		hidden bool
	}{
		{"file1.txt", false, false},
		{"file2.go", false, false},
		{"subdir", true, false},
		{".hidden_file", false, true},
		{".hidden_dir", true, true},
		{"normal_dir", true, false},
		{"README.md", false, false},
	}

	for _, item := range testItems {
		path := filepath.Join(tempDir, item.name)
		if item.isDir {
			err := os.Mkdir(path, 0755)
			if err != nil {
				t.Fatalf("Failed to create directory %s: %v", item.name, err)
			}
		} else {
			err := os.WriteFile(path, []byte("test content"), 0644)
			if err != nil {
				t.Fatalf("Failed to create file %s: %v", item.name, err)
			}
		}
	}

	// Test List function
	result := List(tempDir)

	// Expected items (excluding hidden ones)
	expected := []string{"file1.txt", "file2.go", "subdir", "normal_dir", "README.md"}

	// Check length
	if len(result) != len(expected) {
		t.Errorf("Expected %d items, got %d.\nExpected: %v\nGot: %v", len(expected), len(result), expected, result)
		return
	}

	// Convert to map for easier checking
	resultMap := make(map[string]bool)
	for _, item := range result {
		resultMap[item] = true
	}

	// Check that all expected items are present
	for _, expectedItem := range expected {
		if !resultMap[expectedItem] {
			t.Errorf("Expected item %s not found in result", expectedItem)
		}
	}

	// Check that no hidden items are present
	for _, item := range result {
		if item[0] == '.' {
			t.Errorf("Hidden item %s should not be in result", item)
		}
	}
}

func TestList_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	result := List(tempDir)

	if len(result) != 0 {
		t.Errorf("Expected empty result for empty directory, got %v", result)
	}
}

func TestList_NonExistentDirectory(t *testing.T) {
	nonExistentDir := "/path/that/does/not/exist"

	result := List(nonExistentDir)

	if len(result) != 0 {
		t.Errorf("Expected empty result for non-existent directory, got %v", result)
	}
}

func TestList_HiddenFilesExcluded(t *testing.T) {
	tempDir := t.TempDir()

	// Create various hidden files and directories
	hiddenItems := []string{
		".git",
		".gitignore",
		".hidden_file.txt",
		".DS_Store",
		"..parent_ref",
	}

	for _, item := range hiddenItems {
		path := filepath.Join(tempDir, item)
		if item == ".git" {
			err := os.Mkdir(path, 0755)
			if err != nil {
				t.Fatalf("Failed to create directory %s: %v", item, err)
			}
		} else {
			err := os.WriteFile(path, []byte("hidden content"), 0644)
			if err != nil {
				t.Fatalf("Failed to create file %s: %v", item, err)
			}
		}
	}

	// Also create some normal files
	normalFiles := []string{"normal.txt", "README.md"}
	for _, file := range normalFiles {
		path := filepath.Join(tempDir, file)
		err := os.WriteFile(path, []byte("normal content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	result := List(tempDir)

	// Should only contain normal files
	if len(result) != len(normalFiles) {
		t.Errorf("Expected %d items, got %d.\nExpected: %v\nGot: %v", len(normalFiles), len(result), normalFiles, result)
	}

	// Verify no hidden items are included
	for _, item := range result {
		if item[0] == '.' {
			t.Errorf("Hidden item %s should not be in result", item)
		}
	}

	// Verify all normal files are included
	resultMap := make(map[string]bool)
	for _, item := range result {
		resultMap[item] = true
	}

	for _, file := range normalFiles {
		if !resultMap[file] {
			t.Errorf("Expected normal file %s not found in result", file)
		}
	}
}

func TestList_MixedContent(t *testing.T) {
	tempDir := t.TempDir()

	// Create a mix of files and directories
	items := map[string]bool{
		"file1.txt":   false, // file
		"file2.go":    false, // file
		"directory1":  true,  // directory
		"directory2":  true,  // directory
		"script.sh":   false, // file
		"data":        true,  // directory
		"config.yaml": false, // file
	}

	for name, isDir := range items {
		path := filepath.Join(tempDir, name)
		if isDir {
			err := os.Mkdir(path, 0755)
			if err != nil {
				t.Fatalf("Failed to create directory %s: %v", name, err)
			}
		} else {
			err := os.WriteFile(path, []byte("content"), 0644)
			if err != nil {
				t.Fatalf("Failed to create file %s: %v", name, err)
			}
		}
	}

	result := List(tempDir)

	// Should contain all items
	if len(result) != len(items) {
		t.Errorf("Expected %d items, got %d", len(items), len(result))
	}

	// Verify all items are present
	resultMap := make(map[string]bool)
	for _, item := range result {
		resultMap[item] = true
	}

	for name := range items {
		if !resultMap[name] {
			t.Errorf("Expected item %s not found in result", name)
		}
	}
}

func TestList_PermissionDenied(t *testing.T) {
	// This test might not work on all systems, especially Windows
	// Skip if we can't create a directory with restricted permissions
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tempDir := t.TempDir()
	restrictedDir := filepath.Join(tempDir, "restricted")

	// Create directory
	err := os.Mkdir(restrictedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create restricted directory: %v", err)
	}

	// Remove read permission
	err = os.Chmod(restrictedDir, 0000)
	if err != nil {
		t.Fatalf("Failed to change permissions: %v", err)
	}

	// Restore permissions after test
	defer func() {
		os.Chmod(restrictedDir, 0755)
	}()

	result := List(restrictedDir)

	// Should return empty slice when permission is denied
	if len(result) != 0 {
		t.Errorf("Expected empty result for permission denied directory, got %v", result)
	}
}

func TestList_UnicodeFilenames(t *testing.T) {
	tempDir := t.TempDir()

	// Create files with unicode names
	unicodeFiles := []string{
		"测试文件.txt",
		"файл.go",
		"ファイル.md",
		"🔥emoji_file.txt",
		"café.txt",
	}

	for _, filename := range unicodeFiles {
		path := filepath.Join(tempDir, filename)
		err := os.WriteFile(path, []byte("unicode content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create unicode file %s: %v", filename, err)
		}
	}

	result := List(tempDir)

	if len(result) != len(unicodeFiles) {
		t.Errorf("Expected %d files, got %d", len(unicodeFiles), len(result))
	}

	// Verify all unicode files are present
	resultMap := make(map[string]bool)
	for _, item := range result {
		resultMap[item] = true
	}

	for _, filename := range unicodeFiles {
		if !resultMap[filename] {
			t.Errorf("Expected unicode file %s not found in result", filename)
		}
	}
}
